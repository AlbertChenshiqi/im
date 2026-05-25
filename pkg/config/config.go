package config

import "os"

type Env struct {
	MySQLDSN  string
	RedisAddr string
	JWTSecret string
}

func Load() Env {
	return Env{
		MySQLDSN:  getEnv("MYSQL_DSN", "im:im@tcp(localhost:3306)/im?parseTime=true&charset=utf8mb4&loc=Local"),
		RedisAddr: getEnv("REDIS_ADDR", "localhost:6379"),
		JWTSecret: getEnv("JWT_SECRET", "im-dev-secret-change-in-production"),
	}
}

func getEnv(k, def string) string {
	if v := os.Getenv(k); v != "" {
		return v
	}
	return def
}
