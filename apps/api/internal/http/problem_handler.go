package http

import (
	"log/slog"
	"net/http"

	"github.com/example/oj3/apps/api/internal/problem"
	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
)

func listProblemsHandler(reader problem.Reader, logger *slog.Logger) http.HandlerFunc {
	if logger == nil {
		logger = slog.Default()
	}
	handlerLogger := logger.With("component", "http.problems")

	return func(w http.ResponseWriter, r *http.Request) {
		requestID := middleware.GetReqID(r.Context())
		items, err := reader.List(r.Context())
		if err != nil {
			handlerLogger.ErrorContext(r.Context(), "problem list failed", "event", "problem.list.failed", "requestId", requestID, "error", err)
			writeJSONError(w, http.StatusInternalServerError, "problem list failed")
			return
		}

		response := make([]problemSummaryResponse, 0, len(items))
		for _, item := range items {
			response = append(response, mapProblemSummaryResponse(item))
		}

		if len(items) > 0 {
			w.Header().Set("Last-Modified", items[0].UpdatedAt.UTC().Format(http.TimeFormat))
		}
		handlerLogger.InfoContext(r.Context(), "problem list succeeded", "event", "problem.list.succeeded", "requestId", requestID, "count", len(response))
		writeJSON(w, http.StatusOK, response)
	}
}

func getProblemHandler(reader problem.Reader, logger *slog.Logger) http.HandlerFunc {
	if logger == nil {
		logger = slog.Default()
	}
	handlerLogger := logger.With("component", "http.problems")

	return func(w http.ResponseWriter, r *http.Request) {
		requestID := middleware.GetReqID(r.Context())
		slug := chi.URLParam(r, "slug")
		item, err := reader.GetBySlug(r.Context(), slug)
		if err != nil {
			handlerLogger.ErrorContext(r.Context(), "problem detail failed", "event", "problem.detail.failed", "requestId", requestID, "slug", slug, "error", err)
			writeJSONError(w, http.StatusNotFound, "problem not found")
			return
		}

		w.Header().Set("Last-Modified", item.UpdatedAt.UTC().Format(http.TimeFormat))
		handlerLogger.InfoContext(r.Context(), "problem detail succeeded", "event", "problem.detail.succeeded", "requestId", requestID, "slug", slug)
		writeJSON(w, http.StatusOK, mapProblemDetailResponse(item))
	}
}

func getProblemByRouteCodeHandler(reader problem.Reader, logger *slog.Logger) http.HandlerFunc {
	if logger == nil {
		logger = slog.Default()
	}
	handlerLogger := logger.With("component", "http.problems")

	return func(w http.ResponseWriter, r *http.Request) {
		requestID := middleware.GetReqID(r.Context())
		routeCode := chi.URLParam(r, "routeCode")
		item, err := reader.GetByRouteCode(r.Context(), routeCode)
		if err != nil {
			handlerLogger.ErrorContext(r.Context(), "problem detail by code failed", "event", "problem.detail_by_code.failed", "requestId", requestID, "routeCode", routeCode, "error", err)
			writeJSONError(w, http.StatusNotFound, "problem not found")
			return
		}

		w.Header().Set("Last-Modified", item.UpdatedAt.UTC().Format(http.TimeFormat))
		handlerLogger.InfoContext(r.Context(), "problem detail by code succeeded", "event", "problem.detail_by_code.succeeded", "requestId", requestID, "routeCode", routeCode)
		writeJSON(w, http.StatusOK, mapProblemDetailResponse(item))
	}
}
