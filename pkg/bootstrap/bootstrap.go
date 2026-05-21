package bootstrap

import (
	"context"
	"os"

	"github.com/jackc/pgx/v5/pgxpool"

	"im/pkg/configzero"
	"im/pkg/db"
	"im/pkg/redisclient"
)

func PostgresDSN() string {
	if v := os.Getenv("POSTGRES_DSN"); v != "" {
		return v
	}
	return configzero.PostgresDSN
}

func RedisAddr() string {
	if v := os.Getenv("REDIS_ADDR"); v != "" {
		return v
	}
	return configzero.RedisAddr
}

func JWTSecret() string {
	if v := os.Getenv("JWT_SECRET"); v != "" {
		return v
	}
	return configzero.JWTSecret
}

func NewPGPool(ctx context.Context) (*pgxpool.Pool, error) {
	return db.NewPool(ctx, PostgresDSN())
}

func NewRedis() *redisclient.Client {
	return redisclient.New(RedisAddr())
}
