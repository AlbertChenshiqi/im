package config

import (
	"os"
	"strconv"
)

type Common struct {
	PostgresDSN string
	RedisAddr   string
	KafkaBrokers []string
	JWTSecret   string
	GRPCPort    int
	HTTPPort    int
}

func LoadCommon() Common {
	return Common{
		PostgresDSN: getEnv("POSTGRES_DSN", "postgres://im:im@localhost:5432/im?sslmode=disable"),
		RedisAddr:   getEnv("REDIS_ADDR", "localhost:6379"),
		KafkaBrokers: []string{getEnv("KAFKA_BROKERS", "localhost:9092")},
		JWTSecret:   getEnv("JWT_SECRET", "dev-secret-change-in-production"),
		GRPCPort:    getEnvInt("GRPC_PORT", 50051),
		HTTPPort:    getEnvInt("HTTP_PORT", 8080),
	}
}

func getEnv(key, def string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return def
}

func getEnvInt(key string, def int) int {
	if v := os.Getenv(key); v != "" {
		if n, err := strconv.Atoi(v); err == nil {
			return n
		}
	}
	return def
}

func ServiceAddr(name string, port int) string {
	if v := os.Getenv(name + "_ADDR"); v != "" {
		return v
	}
	return "localhost:" + strconv.Itoa(port)
}
