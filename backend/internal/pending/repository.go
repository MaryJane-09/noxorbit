package pending

import (
	"context"
	"errors"
	"github.com/MaryJane-09/noxorbit/backend/internal/user"
	"github.com/jackc/pgx/v5/pgxpool"
)

type PendingRepository struct {
	pool *pgxpool.Pool
}

func NewRepository(pool *pgxpool.Pool) *PendingRepository {
	return &PendingRepository{
		pool: pool,
	}
}

func (r *PendingRepository) Create(info user.User) error {
	ctx := context.Background()

	_, err := r.pool.Exec(ctx, `INSERT INTO pendings (email, name, password)
	VALUES ($1, $2, $3) ON CONFLICT (email) DO UPDATE
	SET name = $2, password = $3;`, info.Email, info.Name, info.Password)
	if err != nil {
		return err
	}
	return nil
}

func (r *PendingRepository) FindByEmail(email string) (user.User, error) {
	ctx := context.Background()
	row := r.pool.QueryRow(ctx, `SELECT email, name, password FROM pendings WHERE email = $1`, email)
	var p user.User
	err := row.Scan(&p.Email, &p.Name, &p.Password)
	if err != nil {
		return user.User{}, errors.New("Pending user not found")
	}
	return p, nil

}

func (r *PendingRepository) Delete(email string) error {
	ctx := context.Background()

	_, err := r.pool.Exec(ctx, `DELETE FROM pendings WHERE email = $1`, email)
	if err != nil {
		return err
	}
	return nil
}
