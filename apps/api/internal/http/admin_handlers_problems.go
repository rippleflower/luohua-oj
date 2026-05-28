package http

import (
	"encoding/json"
	"net/http"

	"github.com/example/oj3/apps/api/internal/problem"
	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"
)

func adminProblemsHandler(opts authHandlerOptions) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		items, err := opts.service.ListAdminProblems(r.Context())
		if err != nil {
			writeJSONError(w, http.StatusInternalServerError, "failed to load problems")
			return
		}
		response := make([]map[string]any, 0, len(items))
		for _, item := range items {
			response = append(response, adminProblemResponse(problem.AdminProblem{
				ID:               item.ID,
				ProblemNo:        item.ProblemNo,
				RouteCode:        opts.problemRouteCodec.Encode(item.ProblemNo),
				Slug:             item.Slug,
				Title:            item.Title,
				Difficulty:       item.Difficulty,
				TimeLimitMs:      item.TimeLimitMs,
				MemoryLimitKb:    item.MemoryLimitKb,
				Status:           item.Status,
				CurrentVersionNo: item.CurrentVersionNo,
				IsPublished:      item.IsPublished,
				SubmissionCount:  item.SubmissionCount,
				AcceptedRate:     item.AcceptedRate,
				UpdatedAt:        item.UpdatedAt,
			}))
		}
		writeJSON(w, http.StatusOK, response)
	}
}

func adminProblemCreateHandler(opts authHandlerOptions) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		actor, _ := currentUser(r.Context())
		var request struct {
			Slug          string `json:"slug"`
			Title         string `json:"title"`
			Difficulty    string `json:"difficulty"`
			TimeLimitMs   int    `json:"timeLimitMs"`
			MemoryLimitKb int    `json:"memoryLimitKb"`
			Reason        string `json:"reason"`
		}
		if err := json.NewDecoder(r.Body).Decode(&request); err != nil {
			writeJSONError(w, http.StatusBadRequest, "invalid json body")
			return
		}
		item, err := opts.problemAdmin.CreateAdmin(r.Context(), actor, problem.CreateAdminInput{
			Slug:          request.Slug,
			Title:         request.Title,
			Difficulty:    request.Difficulty,
			TimeLimitMs:   request.TimeLimitMs,
			MemoryLimitKb: request.MemoryLimitKb,
			Reason:        request.Reason,
			IP:            requestIP(r),
		})
		if err != nil {
			writeJSONError(w, http.StatusBadRequest, err.Error())
			return
		}
		writeJSON(w, http.StatusCreated, adminProblemResponse(item))
	}
}

func adminProblemUpdateHandler(opts authHandlerOptions) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		actor, _ := currentUser(r.Context())
		problemID, err := uuid.Parse(chi.URLParam(r, "problemID"))
		if err != nil {
			writeJSONError(w, http.StatusBadRequest, "invalid problem id")
			return
		}
		var request struct {
			Slug          string `json:"slug"`
			Title         string `json:"title"`
			Difficulty    string `json:"difficulty"`
			TimeLimitMs   int    `json:"timeLimitMs"`
			MemoryLimitKb int    `json:"memoryLimitKb"`
			Reason        string `json:"reason"`
		}
		if err := json.NewDecoder(r.Body).Decode(&request); err != nil {
			writeJSONError(w, http.StatusBadRequest, "invalid json body")
			return
		}
		item, err := opts.problemAdmin.UpdateAdmin(r.Context(), actor, problem.UpdateAdminInput{
			ProblemID:     problemID,
			Slug:          request.Slug,
			Title:         request.Title,
			Difficulty:    request.Difficulty,
			TimeLimitMs:   request.TimeLimitMs,
			MemoryLimitKb: request.MemoryLimitKb,
			Reason:        request.Reason,
			IP:            requestIP(r),
		})
		if err != nil {
			writeJSONError(w, http.StatusBadRequest, err.Error())
			return
		}
		writeJSON(w, http.StatusOK, adminProblemResponse(item))
	}
}

func adminProblemPublishHandler(opts authHandlerOptions) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		actor, _ := currentUser(r.Context())
		problemID, err := uuid.Parse(chi.URLParam(r, "problemID"))
		if err != nil {
			writeJSONError(w, http.StatusBadRequest, "invalid problem id")
			return
		}
		var request struct {
			Reason string `json:"reason"`
		}
		if err := json.NewDecoder(r.Body).Decode(&request); err != nil {
			writeJSONError(w, http.StatusBadRequest, "invalid json body")
			return
		}
		item, err := opts.problemAdmin.PublishAdmin(r.Context(), actor, problem.PublishAdminInput{
			ProblemID: problemID,
			Reason:    request.Reason,
			IP:        requestIP(r),
		})
		if err != nil {
			writeJSONError(w, http.StatusBadRequest, err.Error())
			return
		}
		writeJSON(w, http.StatusOK, adminProblemResponse(item))
	}
}
