package main

import (
	"context"
	"log/slog"
	"net/http"
	"os"
	"time"

	"github.com/example/oj3/apps/api/internal/auth"
	"github.com/example/oj3/apps/api/internal/contest"
	apphttp "github.com/example/oj3/apps/api/internal/http"
	"github.com/example/oj3/apps/api/internal/platform"
	"github.com/example/oj3/apps/api/internal/problem"
	"github.com/example/oj3/apps/api/internal/queue"
	"github.com/example/oj3/apps/api/internal/source"
	"github.com/example/oj3/apps/api/internal/submission"
	"github.com/jackc/pgx/v5/pgxpool"
)

func main() {
	cfg := platform.LoadConfig()
	logger, logCloser, err := platform.NewLogger(cfg, "api")
	if err != nil {
		slog.Error("logger init failed", "service", "api", "error", err)
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

	judgeQueue := queue.NewClient(cfg.RedisAddr)
	defer judgeQueue.Close()

	submissionService := submission.NewService(
		submission.NewSQLRepository(pool),
		source.LocalStore{Root: cfg.SourceRoot},
		judgeQueue,
		logger.With("component", "submission.service"),
	)
	authService := auth.NewService(
		auth.NewSQLRepository(pool),
		time.Duration(cfg.SessionDurationHours)*time.Hour,
		logger.With("component", "auth.service"),
	)
	problemService := problem.NewService(problem.NewSQLRepository(pool))
	contestService := contest.NewService(contest.NewSQLRepository(pool))

	server := &http.Server{
		Addr: cfg.HTTPAddr,
		Handler: apphttp.NewRouter(apphttp.RouterOptions{
			SubmissionCreator: submissionService,
			SubmissionReader:  submissionService,
			SubmissionAdmin:   submissionService,
			ProblemReader:     problemService,
			ProblemAdmin:      problemService,
			ContestReader:     contestService,
			ContestAdmin:      contestService,
			AuthService:       authService,
			SessionCookieName: cfg.SessionCookieName,
			CookieSecure:      cfg.CookieSecure,
			SourceRoot:        cfg.SourceRoot,
			RedisAddr:         cfg.RedisAddr,
			Logger:            logger,
		}),
	}

	logger.Info("api listening", "component", "bootstrap", "event", "server.started", "addr", cfg.HTTPAddr)
	if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
		logger.Error("api stopped", "component", "bootstrap", "event", "server.stopped", "error", err)
		os.Exit(1)
	}
}
