package platform

import "os"

type Config struct {
	HTTPAddr    string
	DatabaseURL string
	RedisAddr   string
	SourceRoot  string
}

func LoadConfig() Config {
	return Config{
		HTTPAddr:    getenv("HTTP_ADDR", ":8080"),
		DatabaseURL: getenv("DATABASE_URL", "postgres://oj:oj@localhost:5432/oj?sslmode=disable"),
		RedisAddr:   getenv("REDIS_ADDR", "localhost:6379"),
		SourceRoot:  getenv("SOURCE_ROOT", "tmp/submissions"),
	}
}

func getenv(key string, fallback string) string {
	value := os.Getenv(key)
	if value == "" {
		return fallback
	}
	return value
}
