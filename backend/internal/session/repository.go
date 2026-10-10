package session

import (
	"context"
	"errors"
	"time"

	"github.com/gofrs/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

type SessionRepository struct {
	pool *pgxpool.Pool
}

var ErrNotFound = errors.New("session not found")

func NewRepository(pool *pgxpool.Pool) *SessionRepository {
	return &SessionRepository{
		pool: pool,
	}
}

func (r *SessionRepository) Create(userID uuid.UUID, tokenHash string, expiresAt time.Time) error {
	ctx := context.Background()

	_, err := r.pool.Exec(ctx, `INSERT INTO sessions (token_hash, user_id, expires_at)
	VALUES ($1, $2, $3)`, userID, tokenHash, expiresAt)
	if err != nil {
		return err
	}
	return nil
}

func (r *SessionRepository) FindUserID(tokenHash string) (uuid.UUID, error) {
	ctx := context.Background()
	var userID uuid.UUID

	err := r.pool.QueryRow(ctx, `SELECT user_id FROM sessions WHERE token_hash = $1 AND expires_at > now()`, tokenHash).Scan(&userID)

	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return uuid.Nil, ErrNotFound
		}
		return uuid.Nil, err
	}

	return userID, nil
}
