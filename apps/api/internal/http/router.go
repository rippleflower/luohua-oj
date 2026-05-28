package http

import (
	"log/slog"
	"net/http"

	"github.com/example/oj3/apps/api/internal/auth"
	"github.com/example/oj3/apps/api/internal/contest"
	"github.com/example/oj3/apps/api/internal/contest_makeup"
	"github.com/example/oj3/apps/api/internal/problem"
	"github.com/example/oj3/apps/api/internal/submission"
	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
)

type RouterOptions struct {
	SubmissionCreator submission.Creator
	SubmissionReader  submission.Reader
	SubmissionAdmin   submission.AdminManager
	ProblemReader     problem.Reader
	ProblemAdmin      problem.AdminManager
	ContestReader     contest.Reader
	ContestAdmin      contest.AdminManager
	ContestMakeup     contest_makeup.Reader
	AuthService       *auth.Service
	SessionCookieName string
	CookieSecure      bool
	SourceRoot        string
	RedisAddr         string
	ProblemRouteSalt  string
	Logger            *slog.Logger
}

func NewRouter(options ...RouterOptions) http.Handler {
	var opts RouterOptions
	if len(options) > 0 {
		opts = options[0]
	}

	r := chi.NewRouter()
	r.Use(middleware.RequestID)
	r.Use(middleware.RealIP)
	r.Use(middleware.Recoverer)

	r.Get("/health", healthHandler)
	if opts.AuthService != nil {
		authOpts := authHandlerOptions{
			service:           opts.AuthService,
			sessionCookieName: opts.SessionCookieName,
			cookieSecure:      opts.CookieSecure,
			sourceRoot:        opts.SourceRoot,
			redisAddr:         opts.RedisAddr,
			problemAdmin:      opts.ProblemAdmin,
			contestAdmin:      opts.ContestAdmin,
			submissionAdmin:   opts.SubmissionAdmin,
			contestMakeup:     opts.ContestMakeup,
			problemRouteCodec: problem.NewRouteCodec(opts.ProblemRouteSalt),
		}
		mountAuthRoutes(r, authOpts)
		mountAdminRoutes(r, authOpts)
		mountContestMakeupRoutes(r, authOpts)
	}
	if opts.ProblemReader != nil {
		mountProblemRoutes(r, opts.ProblemReader, opts.Logger)
	}
	if opts.ContestReader != nil {
		r.Get("/contests", listContestsHandler(opts.ContestReader, opts.Logger))
		r.Get("/contests/{slug}", getContestHandler(opts.ContestReader, opts.Logger))
	}
	if opts.SubmissionCreator != nil {
		r.Post("/submissions", createSubmissionHandler(opts.SubmissionCreator, opts.Logger))
	}
	if opts.SubmissionReader != nil {
		r.Get("/submissions/{submissionID}", getSubmissionHandler(opts.SubmissionReader, opts.Logger))
		r.Get("/users/{username}/submissions", listUserSubmissionsHandler(opts.SubmissionReader, opts.Logger))
	}

	return r
}
