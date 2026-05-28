package http

import (
	"encoding/json"
	"net/http"
	"time"

	"github.com/example/oj3/apps/api/internal/contest"
	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"
)

func adminContestsHandler(opts authHandlerOptions) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		items, err := opts.contestAdmin.ListAdmin(r.Context())
		if err != nil {
			writeJSONError(w, http.StatusInternalServerError, "failed to load contests")
			return
		}
		response := make([]map[string]any, 0, len(items))
		for _, item := range items {
			response = append(response, adminContestResponse(item))
		}
		writeJSON(w, http.StatusOK, response)
	}
}

func adminContestCreateHandler(opts authHandlerOptions) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		actor, _ := currentUser(r.Context())
		var request struct {
			Slug        string `json:"slug"`
			Title       string `json:"title"`
			Description string `json:"description"`
			Status      string `json:"status"`
			StartsAt    string `json:"startsAt"`
			EndsAt      string `json:"endsAt"`
			Reason      string `json:"reason"`
		}
		if err := json.NewDecoder(r.Body).Decode(&request); err != nil {
			writeJSONError(w, http.StatusBadRequest, "invalid json body")
			return
		}
		startsAt, err := time.Parse(time.RFC3339, request.StartsAt)
		if err != nil {
			writeJSONError(w, http.StatusBadRequest, "invalid startsAt")
			return
		}
		endsAt, err := time.Parse(time.RFC3339, request.EndsAt)
		if err != nil {
			writeJSONError(w, http.StatusBadRequest, "invalid endsAt")
			return
		}
		item, err := opts.contestAdmin.CreateAdmin(r.Context(), actor, contest.CreateAdminInput{
			Slug:        request.Slug,
			Title:       request.Title,
			Description: request.Description,
			Status:      request.Status,
			StartsAt:    startsAt,
			EndsAt:      endsAt,
			Reason:      request.Reason,
			IP:          requestIP(r),
		})
		if err != nil {
			writeJSONError(w, http.StatusBadRequest, err.Error())
			return
		}
		writeJSON(w, http.StatusCreated, adminContestResponse(item))
	}
}

func adminContestUpdateHandler(opts authHandlerOptions) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		actor, _ := currentUser(r.Context())
		contestID, err := uuid.Parse(chi.URLParam(r, "contestID"))
		if err != nil {
			writeJSONError(w, http.StatusBadRequest, "invalid contest id")
			return
		}
		var request struct {
			Slug        string `json:"slug"`
			Title       string `json:"title"`
			Description string `json:"description"`
			Status      string `json:"status"`
			StartsAt    string `json:"startsAt"`
			EndsAt      string `json:"endsAt"`
			Reason      string `json:"reason"`
		}
		if err := json.NewDecoder(r.Body).Decode(&request); err != nil {
			writeJSONError(w, http.StatusBadRequest, "invalid json body")
			return
		}
		startsAt, err := time.Parse(time.RFC3339, request.StartsAt)
		if err != nil {
			writeJSONError(w, http.StatusBadRequest, "invalid startsAt")
			return
		}
		endsAt, err := time.Parse(time.RFC3339, request.EndsAt)
		if err != nil {
			writeJSONError(w, http.StatusBadRequest, "invalid endsAt")
			return
		}
		item, err := opts.contestAdmin.UpdateAdmin(r.Context(), actor, contest.UpdateAdminInput{
			ContestID:   contestID,
			Slug:        request.Slug,
			Title:       request.Title,
			Description: request.Description,
			Status:      request.Status,
			StartsAt:    startsAt,
			EndsAt:      endsAt,
			Reason:      request.Reason,
			IP:          requestIP(r),
		})
		if err != nil {
			writeJSONError(w, http.StatusBadRequest, err.Error())
			return
		}
		writeJSON(w, http.StatusOK, adminContestResponse(item))
	}
}

func adminContestProblemsReplaceHandler(opts authHandlerOptions) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		actor, _ := currentUser(r.Context())
		contestID, err := uuid.Parse(chi.URLParam(r, "contestID"))
		if err != nil {
			writeJSONError(w, http.StatusBadRequest, "invalid contest id")
			return
		}
		var request struct {
			Problems []struct {
				ProblemID string `json:"problemId"`
				Code      string `json:"code"`
				Position  int    `json:"position"`
			} `json:"problems"`
			Reason string `json:"reason"`
		}
		if err := json.NewDecoder(r.Body).Decode(&request); err != nil {
			writeJSONError(w, http.StatusBadRequest, "invalid json body")
			return
		}
		problems := make([]contest.AdminProblemBindingInput, 0, len(request.Problems))
		for _, item := range request.Problems {
			problemID, err := uuid.Parse(item.ProblemID)
			if err != nil {
				writeJSONError(w, http.StatusBadRequest, "invalid problem id")
				return
			}
			problems = append(problems, contest.AdminProblemBindingInput{
				ProblemID: problemID,
				Code:      item.Code,
				Position:  item.Position,
			})
		}
		item, err := opts.contestAdmin.ReplaceProblemsAdmin(r.Context(), actor, contest.ReplaceProblemsInput{
			ContestID: contestID,
			Problems:  problems,
			Reason:    request.Reason,
			IP:        requestIP(r),
		})
		if err != nil {
			writeJSONError(w, http.StatusBadRequest, err.Error())
			return
		}
		writeJSON(w, http.StatusOK, adminContestResponse(item))
	}
}

func adminContestFreezeHandler(opts authHandlerOptions) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		actor, _ := currentUser(r.Context())
		contestID, err := uuid.Parse(chi.URLParam(r, "contestID"))
		if err != nil {
			writeJSONError(w, http.StatusBadRequest, "invalid contest id")
			return
		}
		var request struct {
			Reason string `json:"reason"`
		}
		if err := json.NewDecoder(r.Body).Decode(&request); err != nil {
			writeJSONError(w, http.StatusBadRequest, "invalid json body")
			return
		}
		item, err := opts.contestAdmin.FreezeAdmin(r.Context(), actor, contest.FreezeAdminInput{
			ContestID: contestID,
			Reason:    request.Reason,
			IP:        requestIP(r),
		})
		if err != nil {
			writeJSONError(w, http.StatusBadRequest, err.Error())
			return
		}
		writeJSON(w, http.StatusOK, adminContestResponse(item))
	}
}
