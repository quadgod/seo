package db

import (
	"context"
	"github.com/jackc/pgx/v5/pgxpool"
)

func Connect(ctx context.Context, connectionString string) (*pgxpool.Pool, error) {
	con, err := pgxpool.New(ctx, connectionString)

	if err != nil {
		return nil, err
	}

	return con, nil
}
