package http

import (
	"errors"
	"net/http"

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

		writeJSON(w, http.StatusOK, mapContestMakeupListResponse(list))
	}
}
