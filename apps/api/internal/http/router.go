package http

import (
	"net/http"

	"github.com/example/oj3/apps/api/internal/submission"
	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
)

type RouterOptions struct {
	SubmissionCreator submission.Creator
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
	if opts.SubmissionCreator != nil {
		r.Post("/submissions", createSubmissionHandler(opts.SubmissionCreator))
	}

	return r
}
