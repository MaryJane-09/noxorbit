package otp

import (
	"context"
	"crypto/subtle"
	"errors"
	"time"

	"github.com/gofrs/uuid"
	"golang.org/x/crypto/bcrypt"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

var (
	ErrNotFound        = errors.New("verification code not found")
	ErrExpired         = errors.New("verification code expired")
	ErrInvalidCode     = errors.New("incorrect verification code")
	ErrTooManyAttempts = errors.New("too many verification attempts")
	ErrCooldown        = errors.New("wait before requesting another code")
	ErrAlreadyUsed     = errors.New("verification code already used")
)

type OTPRepository struct {
	pool *pgxpool.Pool
}

func NewRepository(pool *pgxpool.Pool) *OTPRepository {
	return &OTPRepository{pool: pool}
}

// Issue stores a code without storing any account password. The upsert is
// conditional, so simultaneous requests cannot bypass the resend cooldown.
// An accepted new code also resets the guess counter and consumed flag.
func (r *OTPRepository) Issue(code OTP) error {
	ctx := context.Background()
	// An OTP lasts five minutes; its original issue time is derived from
	// expires_at. Only replace it after the first minute has passed.
	cutoff := time.Now().Add(ExpiryTime - IssueCooldown)
	tag, err := r.pool.Exec(ctx, `INSERT INTO otps (email, purpose, code, expires_at, verified, attempts)
		VALUES ($1, $2, $3, $4, false, 0)
		ON CONFLICT (email, purpose) DO UPDATE
		SET code = EXCLUDED.code,
		    expires_at = EXCLUDED.expires_at,
		    verified = false,
		    attempts = 0
		WHERE otps.expires_at <= $5`,
		code.Email, code.Purpose, code.Code, code.ExpiresAt, cutoff)
	if err != nil {
		return err
	}
	if tag.RowsAffected() == 0 {
		return ErrCooldown
	}
	return nil
}

// Verify checks and consumes a one-time code while holding a database row lock.
// Wrong guesses are counted in Postgres; concurrent requests cannot avoid
// the five-attempt cap or redeem the same code twice.
func (r *OTPRepository) Verify(email, purpose, submitted string) error {
	ctx := context.Background()
	tx, err := r.pool.Begin(ctx)
	if err != nil {
		return err
	}
	defer tx.Rollback(ctx)

	var stored string
	var expiresAt time.Time
	var verified bool
	var attempts int
	err = tx.QueryRow(ctx, `SELECT code, expires_at, verified, attempts
		FROM otps WHERE email = $1 AND purpose = $2 FOR UPDATE`,
		email, purpose).Scan(&stored, &expiresAt, &verified, &attempts)
	if errors.Is(err, pgx.ErrNoRows) {
		return ErrNotFound
	}
	if err != nil {
		return err
	}
	wrong, verdict := inspectCode(submitted, stored, expiresAt, verified, attempts, time.Now())
	if verdict != nil && !wrong {
		return verdict
	}
	if wrong {
		attempts++
		if _, err := tx.Exec(ctx, `UPDATE otps SET attempts = $3
			WHERE email = $1 AND purpose = $2`, email, purpose, attempts); err != nil {
			return err
		}
		if err := tx.Commit(ctx); err != nil {
			return err
		}
		return verdict
	}

	if _, err := tx.Exec(ctx, `UPDATE otps SET verified = true
		WHERE email = $1 AND purpose = $2`, email, purpose); err != nil {
		return err
	}
	return tx.Commit(ctx)
}

// inspectCode has no database dependency so the five-guess policy can be
// unit-tested. The caller persists the increment before returning the error.
// wrong is true only when this submission must count as an incorrect guess.
func inspectCode(submitted, stored string, expiresAt time.Time, verified bool, attempts int, now time.Time) (wrong bool, err error) {
	if !now.Before(expiresAt) {
		return false, ErrExpired
	}
	if verified {
		return false, ErrAlreadyUsed
	}
	if attempts >= MaxAttempts {
		return false, ErrTooManyAttempts
	}
	if subtle.ConstantTimeCompare([]byte(submitted), []byte(stored)) != 1 {
		if attempts+1 >= MaxAttempts {
			return true, ErrTooManyAttempts
		}
		return true, ErrInvalidCode
	}
	return false, nil
}

// DeleteIfCode prevents email-send failures from deleting a newer code issued
// by another request.
func (r *OTPRepository) DeleteIfCode(email, purpose, code string) error {
	ctx := context.Background()
	_, err := r.pool.Exec(ctx, `DELETE FROM otps
		WHERE email = $1 AND purpose = $2 AND code = $3`, email, purpose, code)
	return err
}

func (r *OTPRepository) Delete(email, purpose string) error {
	ctx := context.Background()
	_, err := r.pool.Exec(ctx, `DELETE FROM otps
		WHERE email = $1 AND purpose = $2`, email, purpose)
	return err
}

// ResetPassword performs OTP validation, password replacement, session
// revocation, and OTP consumption within one PostgreSQL transaction.
// Invalid guesses are committed separately so the five-guess cap survives
// even though they return an error. A failure on a valid code rolls back
// all changes, leaving both the code and previous sessions intact.
func (r *OTPRepository) ResetPassword(ctx context.Context, email, submitted, newPassword string) error {
	tx, err := r.pool.Begin(ctx)
	if err != nil {
		return err
	}
	defer tx.Rollback(ctx)

	var stored string
	var expiry time.Time
	var used bool
	var attempts int
	err = tx.QueryRow(ctx, `SELECT code, expires_at, verified, attempts
		FROM otps WHERE email = $1 AND purpose = $2 FOR UPDATE`,
		email, PurposeReset).Scan(&stored, &expiry, &used, &attempts)
	if errors.Is(err, pgx.ErrNoRows) {
		return ErrNotFound
	}
	if err != nil {
		return err
	}

	wrong, verdict := inspectCode(submitted, stored, expiry, used, attempts, time.Now())
	if verdict != nil && !wrong {
		return verdict
	}
	if wrong {
		_, err = tx.Exec(ctx, `UPDATE otps SET attempts = $3
			WHERE email = $1 AND purpose = $2`, email, PurposeReset, attempts+1)
		if err != nil {
			return err
		}
		if err := tx.Commit(ctx); err != nil {
			return err
		}
		return verdict
	}

	hash, err := bcrypt.GenerateFromPassword([]byte(newPassword), bcrypt.DefaultCost)
	if err != nil {
		return err
	}

	var userID uuid.UUID
	err = tx.QueryRow(ctx, `UPDATE users SET password = $1
		WHERE email = $2 RETURNING id`, string(hash), email).Scan(&userID)
	if errors.Is(err, pgx.ErrNoRows) {
		return ErrNotFound
	}
	if err != nil {
		return err
	}

	if _, err := tx.Exec(ctx, `DELETE FROM sessions WHERE user_id = $1`, userID); err != nil {
		return err
	}
	tag, err := tx.Exec(ctx, `DELETE FROM otps WHERE email = $1 AND purpose = $2`,
		email, PurposeReset)
	if err != nil {
		return err
	}
	if tag.RowsAffected() != 1 {
		return errors.New("reset code disappeared during transaction")
	}
	return tx.Commit(ctx)
}
