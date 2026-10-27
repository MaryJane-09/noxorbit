package otp

import (
	"context"
	"errors"

	"github.com/jackc/pgx/v5/pgxpool"
)

type OTPRepository struct {
	pool *pgxpool.Pool
}

func NewRepository(pool *pgxpool.Pool) *OTPRepository {
	return &OTPRepository{pool: pool}
}

func (r *OTPRepository) Create(otp OTP) error {
	ctx := context.Background()
	_, err := r.pool.Exec(ctx, `INSERT INTO otps (email, purpose, code, expires_at, verified)
	VALUES ($1, $2, $3, $4, $5)
	ON CONFLICT (email, purpose) DO UPDATE
	SET code = $3, expires_at = $4, verified = $5`,
		otp.Email, otp.Purpose, otp.Code, otp.ExpiresAt, otp.Verified)
	return err
}

func (r *OTPRepository) FindByEmail(email, purpose string) (OTP, error) {
	ctx := context.Background()
	row := r.pool.QueryRow(ctx, `SELECT email, purpose, code, expires_at, verified
	FROM otps WHERE email = $1 AND purpose = $2`, email, purpose)
	var o OTP
	err := row.Scan(&o.Email, &o.Purpose, &o.Code, &o.ExpiresAt, &o.Verified)
	if err != nil {
		return OTP{}, errors.New("OTP not found")
	}
	return o, nil
}

func (r *OTPRepository) Delete(email, purpose string) error {
	ctx := context.Background()
	_, err := r.pool.Exec(ctx, `DELETE FROM otps WHERE email = $1 AND purpose = $2`, email, purpose)
	return err
}
