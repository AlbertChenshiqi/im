package bootstrap

import (
	"database/sql"
	"os"

	"im/pkg/configzero"
	"im/pkg/db"
	"im/pkg/redisclient"
)

func MySQLDSN() string {
	if v := os.Getenv("MYSQL_DSN"); v != "" {
		return v
	}
	return configzero.MySQLDSN
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

func NewMySQL() (*sql.DB, error) {
	return db.NewDB(MySQLDSN())
}

func NewRedis() *redisclient.Client {
	return redisclient.New(RedisAddr())
}
