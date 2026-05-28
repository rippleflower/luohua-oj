package http

import (
	"log/slog"

	"github.com/example/oj3/apps/api/internal/problem"
	"github.com/go-chi/chi/v5"
)

func mountProblemRoutes(r chi.Router, reader problem.Reader, logger *slog.Logger) {
	r.Get("/problems", listProblemsHandler(reader, logger))
	r.Get("/problems/code/{routeCode}", getProblemByRouteCodeHandler(reader, logger))
	r.Get("/problems/{slug}", getProblemHandler(reader, logger))
}
