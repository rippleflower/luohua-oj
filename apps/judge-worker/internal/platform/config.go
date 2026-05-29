package platform

import (
	"os"
	"strconv"
)

type Config struct {
	DatabaseURL       string
	RedisAddr         string
	SourceRoot        string
	WorkRoot          string
	SandboxBin        string
	LogLevel          string
	LogDir            string
	LogFileName       string
	LogFormat         string
	ProblemRouteSalt  string
	WorkerConcurrency int
}

func LoadConfig() Config {
	return Config{
		DatabaseURL:       getenv("DATABASE_URL", "postgres://oj:oj@localhost:5432/oj?sslmode=disable"),
		RedisAddr:         getenv("REDIS_ADDR", "localhost:6379"),
		SourceRoot:        getenv("SOURCE_ROOT", "tmp/submissions"),
		WorkRoot:          getenv("WORK_ROOT", "tmp/judge-runs"),
		SandboxBin:        getenv("SANDBOX_BINARY", "nsjail"),
		LogLevel:          getenv("LOG_LEVEL", "info"),
		LogDir:            getenv("LOG_DIR", "tmp/logs"),
		LogFileName:       getenv("LOG_FILE_NAME", "judge-worker.log"),
		LogFormat:         getenv("LOG_FORMAT", "json"),
		ProblemRouteSalt:  getenv("PROBLEM_ROUTE_SALT", ""),
		WorkerConcurrency: getenvInt("JUDGE_WORKER_CONCURRENCY", 4),
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
