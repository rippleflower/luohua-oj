package http

import (
	"encoding/json"
	"net/http"
	"time"

	"github.com/example/oj3/apps/api/internal/submission"
	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"
)

func adminSubmissionsHandler(opts authHandlerOptions) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		items, err := opts.service.ListAdminSubmissions(r.Context())
		if err != nil {
			writeJSONError(w, http.StatusInternalServerError, "failed to load submissions")
			return
		}
		response := make([]map[string]any, 0, len(items))
		for _, item := range items {
			response = append(response, map[string]any{
				"id":           item.ID.String(),
				"username":     item.Username,
				"problemTitle": item.ProblemTitle,
				"language":     item.Language,
				"status":       item.Status,
				"createdAt":    item.CreatedAt.UTC().Format(time.RFC3339),
			})
		}
		writeJSON(w, http.StatusOK, response)
	}
}

func adminJudgeQueueHandler(opts authHandlerOptions) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		summary, err := opts.submissionAdmin.QueueSummary(r.Context())
		if err != nil {
			writeJSONError(w, http.StatusInternalServerError, "failed to load judge queue")
			return
		}
		writeJSON(w, http.StatusOK, judgeQueueSummaryResponse(summary))
	}
}

func adminSubmissionRejudgeHandler(opts authHandlerOptions) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		actor, _ := currentUser(r.Context())
		submissionID, err := uuid.Parse(chi.URLParam(r, "submissionID"))
		if err != nil {
			writeJSONError(w, http.StatusBadRequest, "invalid submission id")
			return
		}
		var request struct {
			Reason string `json:"reason"`
		}
		if err := json.NewDecoder(r.Body).Decode(&request); err != nil {
			writeJSONError(w, http.StatusBadRequest, "invalid json body")
			return
		}
		result, err := opts.submissionAdmin.RejudgeAdmin(r.Context(), actor, submission.RejudgeAdminInput{
			SubmissionID: submissionID,
			Reason:       request.Reason,
			IP:           requestIP(r),
		})
		if err != nil {
			writeJSONError(w, http.StatusBadRequest, err.Error())
			return
		}
		writeJSON(w, http.StatusOK, map[string]any{
			"submissionId":          result.SubmissionID.String(),
			"queue":                 result.Queue,
			"resultSnapshotVersion": result.ResultSnapshotVersion,
			"problemVersionId":      nullableUUIDStringPtr(result.ProblemVersionID),
		})
	}
}
