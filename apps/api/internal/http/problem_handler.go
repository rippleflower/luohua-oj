package http

import (
	"encoding/json"
	"log/slog"
	"net/http"
	"time"

	"github.com/example/oj3/apps/api/internal/problem"
	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
)

type problemSummaryResponse struct {
	ID           string   `json:"id"`
	Slug         string   `json:"slug"`
	Title        string   `json:"title"`
	Difficulty   string   `json:"difficulty"`
	Tags         []string `json:"tags"`
	AcceptedRate float64  `json:"acceptedRate"`
}

type problemDetailResponse struct {
	ID            string          `json:"id"`
	Slug          string          `json:"slug"`
	Title         string          `json:"title"`
	Difficulty    string          `json:"difficulty"`
	StatementJSON json.RawMessage `json:"statementJson"`
	SamplesJSON   json.RawMessage `json:"samplesJson"`
	LimitsJSON    json.RawMessage `json:"limitsJson"`
	MetadataJSON  json.RawMessage `json:"metadataJson"`
	UpdatedAt     string          `json:"updatedAt"`
}

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
			response = append(response, problemSummaryResponse{
				ID:           item.ID.String(),
				Slug:         item.Slug,
				Title:        item.Title,
				Difficulty:   item.Difficulty,
				Tags:         item.Tags,
				AcceptedRate: item.AcceptedRate,
			})
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
		writeJSON(w, http.StatusOK, problemDetailResponse{
			ID:            item.ID.String(),
			Slug:          item.Slug,
			Title:         item.Title,
			Difficulty:    item.Difficulty,
			StatementJSON: item.StatementJSON,
			SamplesJSON:   item.SamplesJSON,
			LimitsJSON:    item.LimitsJSON,
			MetadataJSON:  item.MetadataJSON,
			UpdatedAt:     item.UpdatedAt.UTC().Format(time.RFC3339),
		})
	}
}
