package otp

import (
	"context"
	"crypto/subtle"
	"errors"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

var (
	ErrNotFound        = errors.New("verification code not found")
	ErrExpired         = errors.New("verification code expired")
	ErrInvalidCode     = errors.New("incorrect verification code")
	ErrTooManyAttempts = errors.New("too many verification attempts")
	ErrCooldown        = errors.New("wait before requesting another code")
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
	if !time.Now().Before(expiresAt) {
		return ErrExpired
	}
	if verified || attempts >= MaxAttempts {
		return ErrTooManyAttempts
	}

	if subtle.ConstantTimeCompare([]byte(submitted), []byte(stored)) != 1 {
		attempts++
		if _, err := tx.Exec(ctx, `UPDATE otps SET attempts = $3
			WHERE email = $1 AND purpose = $2`, email, purpose, attempts); err != nil {
			return err
		}
		if err := tx.Commit(ctx); err != nil {
			return err
		}
		if attempts >= MaxAttempts {
			return ErrTooManyAttempts
		}
		return ErrInvalidCode
	}

	if _, err := tx.Exec(ctx, `UPDATE otps SET verified = true
		WHERE email = $1 AND purpose = $2`, email, purpose); err != nil {
		return err
	}
	return tx.Commit(ctx)
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
