package repo

import (
	"context"

	"github.com/jackc/pgx/v5/pgxpool"

	"im/pkg/db"
)

func NewPool(ctx context.Context, dsn string) (*pgxpool.Pool, error) {
	return db.NewPool(ctx, dsn)
}
