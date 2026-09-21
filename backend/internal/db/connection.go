package db

import (
	"context"
	"github.com/MaryJane-09/noxorbit/backend/config"
	"github.com/jackc/pgx/v5/pgxpool"
)

func Connect() (*pgxpool.Pool, error){
	connString := config.AppConfig.DatabaseURL
	ctx := context.Background()
	pool, err := pgxpool.New(ctx, connString)
	if err != nil{
		return nil, err
	}
	return pool, nil
}
