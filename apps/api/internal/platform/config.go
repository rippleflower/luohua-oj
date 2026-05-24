package platform

import (
	"os"
	"strconv"
)

type Config struct {
	HTTPAddr             string
	DatabaseURL          string
	RedisAddr            string
	SourceRoot           string
	LogLevel             string
	LogDir               string
	LogFileName          string
	LogFormat            string
	SessionCookieName    string
	SessionDurationHours int
	CookieSecure         bool
}

func LoadConfig() Config {
	return Config{
		HTTPAddr:             getenv("HTTP_ADDR", ":8080"),
		DatabaseURL:          getenv("DATABASE_URL", "postgres://oj:oj@localhost:5432/oj?sslmode=disable"),
		RedisAddr:            getenv("REDIS_ADDR", "localhost:6379"),
		SourceRoot:           getenv("SOURCE_ROOT", "tmp/submissions"),
		LogLevel:             getenv("LOG_LEVEL", "info"),
		LogDir:               getenv("LOG_DIR", "tmp/logs"),
		LogFileName:          getenv("LOG_FILE_NAME", "api.log"),
		LogFormat:            getenv("LOG_FORMAT", "json"),
		SessionCookieName:    getenv("SESSION_COOKIE_NAME", "oj_session"),
		SessionDurationHours: getenvInt("SESSION_DURATION_HOURS", 336),
		CookieSecure:         getenv("COOKIE_SECURE", "false") == "true",
	}
}

func getenv(key string, fallback string) string {
	value := os.Getenv(key)
	if value == "" {
		return fallback
	}
	return value
}

func getenvInt(key string, fallback int) int {
	value := os.Getenv(key)
	if value == "" {
		return fallback
	}

	parsed, err := strconv.Atoi(value)
	if err != nil {
		return fallback
	}

	return parsed
}
