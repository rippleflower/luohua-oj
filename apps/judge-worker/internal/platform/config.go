package platform

import "os"

type Config struct {
	DatabaseURL string
	RedisAddr   string
}

func LoadConfig() Config {
	return Config{
		DatabaseURL: getenv("DATABASE_URL", "postgres://oj:oj@localhost:5432/oj?sslmode=disable"),
		RedisAddr:   getenv("REDIS_ADDR", "localhost:6379"),
	}
}

func getenv(key string, fallback string) string {
	value := os.Getenv(key)
	if value == "" {
		return fallback
	}
	return value
}
