package platform

import "os"

type Config struct {
	DatabaseURL string
	RedisAddr   string
	SourceRoot  string
	WorkRoot    string
	SandboxBin  string
	LogLevel    string
	LogDir      string
	LogFileName string
	LogFormat   string
}

func LoadConfig() Config {
	return Config{
		DatabaseURL: getenv("DATABASE_URL", "postgres://oj:oj@localhost:5432/oj?sslmode=disable"),
		RedisAddr:   getenv("REDIS_ADDR", "localhost:6379"),
		SourceRoot:  getenv("SOURCE_ROOT", "tmp/submissions"),
		WorkRoot:    getenv("WORK_ROOT", "tmp/judge-runs"),
		SandboxBin:  getenv("SANDBOX_BINARY", "nsjail"),
		LogLevel:    getenv("LOG_LEVEL", "info"),
		LogDir:      getenv("LOG_DIR", "tmp/logs"),
		LogFileName: getenv("LOG_FILE_NAME", "judge-worker.log"),
		LogFormat:   getenv("LOG_FORMAT", "json"),
	}
}

func getenv(key string, fallback string) string {
	value := os.Getenv(key)
	if value == "" {
		return fallback
	}
	return value
}
