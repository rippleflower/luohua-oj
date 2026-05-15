package main

import (
	"context"
	"net/http"
	"os"

	db "github.com/example/oj3/apps/api/internal/db/generated"
	apphttp "github.com/example/oj3/apps/api/internal/http"
	"github.com/example/oj3/apps/api/internal/platform"
	"github.com/example/oj3/apps/api/internal/queue"
	"github.com/example/oj3/apps/api/internal/source"
	"github.com/example/oj3/apps/api/internal/submission"
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

	judgeQueue := queue.NewClient(cfg.RedisAddr)
	defer judgeQueue.Close()

	submissionService := submission.NewService(
		submission.NewSQLRepository(db.New(pool)),
		source.LocalStore{Root: cfg.SourceRoot},
		judgeQueue,
	)

	server := &http.Server{
		Addr:    cfg.HTTPAddr,
		Handler: apphttp.NewRouter(apphttp.RouterOptions{SubmissionCreator: submissionService}),
	}

	logger.Info("api listening", "addr", cfg.HTTPAddr)
	if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
		logger.Error("api stopped", "error", err)
		os.Exit(1)
	}
}
