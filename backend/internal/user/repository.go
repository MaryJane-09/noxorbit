package user

import (
	"context"
	"errors"
	"github.com/gofrs/uuid"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/jackc/pgx/v5/pgxpool"
	"golang.org/x/crypto/bcrypt"
	"github.com/jackc/pgx/v5"
)

type UserRepository struct {
	pool  *pgxpool.Pool
}

func NewRepository(pool *pgxpool.Pool) *UserRepository {
	return &UserRepository{
		pool:  pool,
	}
}
var ErrEmailExists = errors.New("this email already exists")
var ErrUserNotFound = errors.New("User not found")


func (r *UserRepository) Create(user User) error {
	ctx := context.Background()
	hash := []byte(user.Password)
	hashed, err := bcrypt.GenerateFromPassword(hash, bcrypt.DefaultCost)
	if err != nil {
		return err
	}
	user.Password = string(hashed)

	id, err := uuid.NewV4()
	if err != nil {
		return err
	}
	user.Id = id

	_, err = r.pool.Exec(ctx, `INSERT INTO users (id, name, email, password)
	VALUES ($1, $2, $3, $4)`, user.Id, user.Name, user.Email, user.Password)
	if err != nil {
		var pgErr *pgconn.PgError
		if errors.As(err, &pgErr) {
			if pgErr.Code == "23505" {
				return ErrEmailExists
			}
		}
		return err
	}
	return nil
}

func (r *UserRepository) FindByID(id uuid.UUID) (User, error) {
	ctx := context.Background()
	row := r.pool.QueryRow(ctx, `SELECT id, name, email, password FROM users WHERE id = $1`, id)
	var u User
	err := row.Scan(&u.Id, &u.Name, &u.Email, &u.Password)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return User{}, ErrUserNotFound
		}
		return User{}, err
	}
	return u, nil
}

func (r *UserRepository) FindByEmail(email string) (User, error) {

	ctx := context.Background()
	row := r.pool.QueryRow(ctx, `SELECT id, name, email, password FROM users WHERE email = $1`, email)
	var u User
	err := row.Scan(&u.Id, &u.Name, &u.Email, &u.Password)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return User{}, ErrUserNotFound
		}
		return User{}, err
	}
	return u, nil

}

func (r *UserRepository) GetAllUsers() []UserResponse {
	ctx := context.Background()
	rows, err := r.pool.Query(ctx, `SELECT name, email FROM users `)
	if err != nil {
		return []UserResponse{}
	}
	defer rows.Close()
	var users []UserResponse

	for rows.Next() {
		var u UserResponse
		err = rows.Scan(&u.Name, &u.Email)
		if err != nil {
			return users
		}
		users = append(users, u)
	}
	return users
}
