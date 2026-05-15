package main

import (
	"context"
	"os"

	"github.com/example/oj3/apps/judge-worker/internal/judge"
	"github.com/example/oj3/apps/judge-worker/internal/platform"
	"github.com/hibiken/asynq"
	"github.com/jackc/pgx/v5/pgxpool"
)

func main() {
	cfg := platform.LoadConfig()
	logger := platform.NewLogger()
	ctx := context.Background()

	pool, err := pgxpool.New(ctx, cfg.DatabaseURL)
	if err != nil {
		logger.Error("database connection failed", "error", err)
		os.Exit(1)
	}
	defer pool.Close()

	server := asynq.NewServer(
		asynq.RedisClientOpt{Addr: cfg.RedisAddr},
		asynq.Config{
			Concurrency: 2,
			Queues: map[string]int{
				"judge": 1,
			},
		},
	)

	mux := asynq.NewServeMux()
	judge.NewProcessor(judge.NewSQLRepository(pool), judge.NotImplementedExecutor{}).Register(mux)

	logger.Info("judge worker listening")
	if err := server.Run(mux); err != nil {
		logger.Error("judge worker stopped", "error", err)
		os.Exit(1)
	}
}
