package otp

import (
	"context"
	"errors"
	"fmt"
	"os"
	"strings"
	"sync"
	"testing"
	"time"

	"github.com/gofrs/uuid"
	"github.com/jackc/pgx/v5/pgxpool"
	"golang.org/x/crypto/bcrypt"
)

// Opt-in integration test. NEVER point this at the production database.
// This test creates and destroys its own randomly named PostgreSQL schema.
// Run only with NOXORBIT_INTEGRATION_DB_ACK=disposable and a dedicated
// NOXORBIT_INTEGRATION_TEST_DATABASE_URL.
func testIsolatedPool(t *testing.T) *pgxpool.Pool {
	t.Helper()
	url := os.Getenv("NOXORBIT_INTEGRATION_TEST_DATABASE_URL")
	if url == "" || os.Getenv("NOXORBIT_INTEGRATION_DB_ACK") != "disposable" {
		t.Skip("requires an explicitly acknowledged disposable PostgreSQL test database")
	}
	ctx := context.Background()
	admin, err := pgxpool.New(ctx, url)
	if err != nil { t.Fatal(err) }
	t.Cleanup(admin.Close)

	u, err := uuid.NewV4()
	if err != nil { t.Fatal(err) }
	schema := "nox_test_" + strings.ReplaceAll(u.String(), "-", "")
	if _, err := admin.Exec(ctx, "CREATE SCHEMA "+schema); err != nil { t.Fatal(err) }
	t.Cleanup(func() { _, _ = admin.Exec(context.Background(), "DROP SCHEMA "+schema+" CASCADE") })

	ddl := []string{
		"CREATE TABLE " + schema + ".users (id UUID PRIMARY KEY, email TEXT UNIQUE NOT NULL, password TEXT NOT NULL)",
		"CREATE TABLE " + schema + ".sessions (token_hash TEXT PRIMARY KEY, user_id UUID NOT NULL)",
		"CREATE TABLE " + schema + ".otps (email TEXT NOT NULL, purpose TEXT NOT NULL, code TEXT NOT NULL, expires_at TIMESTAMPTZ NOT NULL, verified BOOLEAN NOT NULL DEFAULT false, attempts INTEGER NOT NULL DEFAULT 0, PRIMARY KEY(email,purpose))",
	}
	for _, statement := range ddl {
		if _, err := admin.Exec(ctx, statement); err != nil { t.Fatal(err) }
	}
	cfg, err := pgxpool.ParseConfig(url)
	if err != nil { t.Fatal(err) }
	cfg.ConnConfig.RuntimeParams["search_path"] = schema
	cfg.MaxConns = 12
	pool, err := pgxpool.NewWithConfig(ctx, cfg)
	if err != nil { t.Fatal(err) }
	t.Cleanup(pool.Close)
	if err := pool.Ping(ctx); err != nil { t.Fatal(err) }
	return pool
}

func TestConcurrentGuessLimitIntegration(t *testing.T) {
	pool := testIsolatedPool(t)
	ctx := context.Background()
	email := "concurrent@example.test"
	_, err := pool.Exec(ctx, `INSERT INTO otps (email,purpose,code,expires_at)
		VALUES ($1,$2,$3,$4)`, email, PurposeRegister, "abc234", time.Now().Add(5*time.Minute))
	if err != nil { t.Fatal(err) }

	repository := NewRepository(pool)
	var wg sync.WaitGroup
	errs := make(chan error, 10)
	for i := 0; i < 10; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			errs <- repository.Verify(email, PurposeRegister, "wrong1")
		}()
	}
	wg.Wait()
	close(errs)
	for e := range errs {
		if !errors.Is(e, ErrInvalidCode) && !errors.Is(e, ErrTooManyAttempts) {
			t.Fatalf("unexpected verification result: %v", e)
		}
	}
	var attempts int
	if err := pool.QueryRow(ctx, `SELECT attempts FROM otps
		WHERE email = $1 AND purpose = $2`, email, PurposeRegister).Scan(&attempts); err != nil { t.Fatal(err) }
	if attempts != MaxAttempts { t.Fatalf("concurrent attempts = %d, want %d", attempts, MaxAttempts) }
}

func TestResetTransactionRollsBackAndRevokesSessionsIntegration(t *testing.T) {
	pool := testIsolatedPool(t)
	ctx := context.Background()
	repository := NewRepository(pool)
	userID, err := uuid.NewV4()
	if err != nil { t.Fatal(err) }
	email := "reset@example.test"
	oldHash, err := bcrypt.GenerateFromPassword([]byte("OriginalPassword123"), bcrypt.DefaultCost)
	if err != nil { t.Fatal(err) }
	if _, err := pool.Exec(ctx, `INSERT INTO users(id,email,password) VALUES ($1,$2,$3)`, userID, email, string(oldHash)); err != nil { t.Fatal(err) }
	if _, err := pool.Exec(ctx, `INSERT INTO sessions(token_hash,user_id) VALUES ($1,$2)`, "old-session", userID); err != nil { t.Fatal(err) }
	if _, err := pool.Exec(ctx, `INSERT INTO otps(email,purpose,code,expires_at)
		VALUES ($1,$2,$3,$4)`, email, PurposeReset, "abc234", time.Now().Add(5*time.Minute)); err != nil { t.Fatal(err) }

	// Deliberately make session revocation fail. The transaction must
	// preserve the prior password, reset code and existing session.
	var schema string
	if err := pool.QueryRow(ctx, "SELECT current_schema()").Scan(&schema); err != nil { t.Fatal(err) }
	fn := fmt.Sprintf(`CREATE FUNCTION %s.fail_session_delete() RETURNS trigger
		LANGUAGE plpgsql AS $$ BEGIN RAISE EXCEPTION 'deliberate integration failure'; END; $$`, schema)
	if _, err := pool.Exec(ctx, fn); err != nil { t.Fatal(err) }
	if _, err := pool.Exec(ctx, "CREATE TRIGGER fail_session_delete BEFORE DELETE ON sessions FOR EACH ROW EXECUTE FUNCTION "+schema+".fail_session_delete()"); err != nil { t.Fatal(err) }

	if err := repository.ResetPassword(ctx, email, "abc234", "NewPassword1234"); err == nil {
		t.Fatal("reset should fail while session revocation is blocked")
	}
	var storedHash string
	if err := pool.QueryRow(ctx, "SELECT password FROM users WHERE id = $1", userID).Scan(&storedHash); err != nil { t.Fatal(err) }
	if err := bcrypt.CompareHashAndPassword([]byte(storedHash), []byte("OriginalPassword123")); err != nil {
		t.Fatal("failed reset changed the password", err)
	}
	var count int
	if err := pool.QueryRow(ctx, "SELECT COUNT(*) FROM sessions WHERE user_id = $1", userID).Scan(&count); err != nil { t.Fatal(err) }
	if count != 1 { t.Fatalf("failed reset left %d sessions, want 1", count) }
	if err := pool.QueryRow(ctx, "SELECT COUNT(*) FROM otps WHERE email = $1 AND purpose = $2", email, PurposeReset).Scan(&count); err != nil { t.Fatal(err) }
	if count != 1 { t.Fatalf("failed reset left %d OTP rows, want 1", count) }

	if _, err := pool.Exec(ctx, "DROP TRIGGER fail_session_delete ON sessions"); err != nil { t.Fatal(err) }
	if err := repository.ResetPassword(ctx, email, "abc234", "NewPassword1234"); err != nil { t.Fatal(err) }
	if err := pool.QueryRow(ctx, "SELECT password FROM users WHERE id = $1", userID).Scan(&storedHash); err != nil { t.Fatal(err) }
	if err := bcrypt.CompareHashAndPassword([]byte(storedHash), []byte("NewPassword1234")); err != nil { t.Fatal("new password not installed", err) }
	if err := pool.QueryRow(ctx, "SELECT COUNT(*) FROM sessions WHERE user_id = $1", userID).Scan(&count); err != nil { t.Fatal(err) }
	if count != 0 { t.Fatalf("sessions after successful reset = %d", count) }
	if err := pool.QueryRow(ctx, "SELECT COUNT(*) FROM otps WHERE email = $1 AND purpose = $2", email, PurposeReset).Scan(&count); err != nil { t.Fatal(err) }
	if count != 0 { t.Fatalf("codes after successful reset = %d", count) }
}
