package http

import (
	"errors"
	"net/http"
	"time"

	"github.com/example/oj3/apps/api/internal/contest_makeup"
	"github.com/go-chi/chi/v5"
)

func contestMakeupListHandler(opts authHandlerOptions) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		actor, ok := currentUser(r.Context())
		if !ok {
			writeJSONError(w, http.StatusUnauthorized, "authentication required")
			return
		}
		slug := chi.URLParam(r, "slug")
		list, err := opts.contestMakeup.GetMakeupList(r.Context(), actor, slug)
		if err != nil {
			switch {
			case errors.Is(err, contest_makeup.ErrContestNotFound):
				writeJSONError(w, http.StatusNotFound, "contest not found")
			case errors.Is(err, contest_makeup.ErrContestNotEnded):
				writeJSONError(w, http.StatusConflict, "contest is not ended")
			default:
				writeJSONError(w, http.StatusInternalServerError, "failed to build makeup list")
			}
			return
		}

		responseItems := make([]map[string]any, 0, len(list.Items))
		for _, item := range list.Items {
			response := map[string]any{
				"problemId":       item.ProblemID.String(),
				"problemCode":     item.ProblemCode,
				"problemSlug":     item.ProblemSlug,
				"problemTitle":    item.ProblemTitle,
				"difficulty":      item.Difficulty,
				"category":        item.Category,
				"lastStatus":      item.LastStatus,
				"severityRank":    item.SeverityRank,
				"reasonSummary":   item.ReasonSummary,
				"suggestedAction": item.SuggestedAction,
			}
			if item.AttemptCount > 0 {
				response["attemptCount"] = item.AttemptCount
			}
			responseItems = append(responseItems, response)
		}
		writeJSON(w, http.StatusOK, map[string]any{
			"contestSlug": list.ContestSlug,
			"generatedAt": list.GeneratedAt.UTC().Format(time.RFC3339),
			"items":       responseItems,
		})
	}
}
