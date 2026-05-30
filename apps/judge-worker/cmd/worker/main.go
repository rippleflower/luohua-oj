package main

import (
	"context"
	"log/slog"
	"os"

	"github.com/example/oj3/apps/judge-worker/internal/judge"
	"github.com/example/oj3/apps/judge-worker/internal/platform"
	"github.com/example/oj3/apps/judge-worker/internal/sandbox"
	"github.com/hibiken/asynq"
	"github.com/jackc/pgx/v5/pgxpool"
)

func main() {
	cfg := platform.LoadConfig()
	logger, logCloser, err := platform.NewLogger(cfg, "judge-worker")
	if err != nil {
		slog.Error("logger init failed", "service", "judge-worker", "error", err)
		os.Exit(1)
	}
	defer func() {
		_ = logCloser.Close()
	}()
	ctx := context.Background()

	pool, err := pgxpool.New(ctx, cfg.DatabaseURL)
	if err != nil {
		logger.Error("database connection failed", "component", "bootstrap", "event", "db.connect.failed", "error", err)
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
	judge.NewProcessor(
		judge.NewSQLRepository(pool),
		judge.LocalExecutor{
			SourceRoot: cfg.SourceRoot,
			WorkRoot:   cfg.WorkRoot,
			Runner: sandbox.Runner{
				Binary: cfg.SandboxBin,
			},
		},
		logger.With("component", "judge.processor"),
	).Register(mux)

	logger.Info("judge worker listening", "component", "bootstrap", "event", "worker.started")
	if err := server.Run(mux); err != nil {
		logger.Error("judge worker stopped", "component", "bootstrap", "event", "worker.stopped", "error", err)
		os.Exit(1)
	}
}
