package zerokit

import (
	"context"
	"log"

	"github.com/jackc/pgx/v5/pgxpool"

	"im/pkg/repo"
)

func MustPGPool(ctx context.Context, dsn string) *pgxpool.Pool {
	pool, err := repo.NewPool(ctx, dsn)
	if err != nil {
		log.Fatalf("postgres: %v", err)
	}
	return pool
}
