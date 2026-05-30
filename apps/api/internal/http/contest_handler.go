package http

import (
	"log/slog"
	"net/http"
	"time"

	"github.com/example/oj3/apps/api/internal/contest"
	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
)

type contestSummaryResponse struct {
	ID               string `json:"id"`
	Slug             string `json:"slug"`
	Title            string `json:"title"`
	Status           string `json:"status"`
	StartsAt         string `json:"startsAt"`
	EndsAt           string `json:"endsAt"`
	Duration         string `json:"duration"`
	ProblemCount     int    `json:"problemCount"`
	ParticipantCount int    `json:"participantCount"`
	Blurb            string `json:"blurb"`
}

type contestDetailResponse struct {
	contestSummaryResponse
	RankSummary       string                     `json:"rankSummary"`
	Remaining         string                     `json:"remaining"`
	RecentSubmissions []contest.RecentSubmission `json:"recentSubmissions"`
	Problems          []contest.ProblemSnapshot  `json:"problems"`
}

func listContestsHandler(reader contest.Reader, logger *slog.Logger) http.HandlerFunc {
	if logger == nil {
		logger = slog.Default()
	}
	handlerLogger := logger.With("component", "http.contests")

	return func(w http.ResponseWriter, r *http.Request) {
		requestID := middleware.GetReqID(r.Context())
		items, err := reader.List(r.Context())
		if err != nil {
			handlerLogger.ErrorContext(r.Context(), "contest list failed", "event", "contest.list.failed", "requestId", requestID, "error", err)
			writeJSONError(w, http.StatusInternalServerError, "contest list failed")
			return
		}

		response := make([]contestSummaryResponse, 0, len(items))
		for _, item := range items {
			response = append(response, contestSummaryResponse{
				ID:               item.ID.String(),
				Slug:             item.Slug,
				Title:            item.Title,
				Status:           item.Status,
				StartsAt:         item.StartsAt.UTC().Format(time.RFC3339),
				EndsAt:           item.EndsAt.UTC().Format(time.RFC3339),
				Duration:         item.DurationLabel,
				ProblemCount:     item.ProblemCount,
				ParticipantCount: item.ParticipantCount,
				Blurb:            item.Blurb,
			})
		}

		handlerLogger.InfoContext(r.Context(), "contest list succeeded", "event", "contest.list.succeeded", "requestId", requestID, "count", len(response))
		writeJSON(w, http.StatusOK, response)
	}
}

func getContestHandler(reader contest.Reader, logger *slog.Logger) http.HandlerFunc {
	if logger == nil {
		logger = slog.Default()
	}
	handlerLogger := logger.With("component", "http.contests")

	return func(w http.ResponseWriter, r *http.Request) {
		requestID := middleware.GetReqID(r.Context())
		slug := chi.URLParam(r, "slug")
		item, err := reader.GetBySlug(r.Context(), slug)
		if err != nil {
			handlerLogger.ErrorContext(r.Context(), "contest detail failed", "event", "contest.detail.failed", "requestId", requestID, "slug", slug, "error", err)
			writeJSONError(w, http.StatusNotFound, "contest not found")
			return
		}

		w.Header().Set("Last-Modified", item.UpdatedAt.UTC().Format(http.TimeFormat))
		handlerLogger.InfoContext(r.Context(), "contest detail succeeded", "event", "contest.detail.succeeded", "requestId", requestID, "slug", slug)
		writeJSON(w, http.StatusOK, contestDetailResponse{
			contestSummaryResponse: contestSummaryResponse{
				ID:               item.ID.String(),
				Slug:             item.Slug,
				Title:            item.Title,
				Status:           item.Status,
				StartsAt:         item.StartsAt.UTC().Format(time.RFC3339),
				EndsAt:           item.EndsAt.UTC().Format(time.RFC3339),
				Duration:         item.DurationLabel,
				ProblemCount:     item.ProblemCount,
				ParticipantCount: item.ParticipantCount,
				Blurb:            item.Blurb,
			},
			RankSummary:       item.RankSummary,
			Remaining:         item.Remaining,
			RecentSubmissions: item.RecentSubmissions,
			Problems:          item.Problems,
		})
	}
}
