package http

import (
	"net/http"

	"github.com/example/oj3/apps/api/internal/contest"
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
		request, err := decodeAdminContestUpsertRequest(r)
		if err != nil {
			writeJSONError(w, http.StatusBadRequest, "invalid json body")
			return
		}
		startsAt, endsAt, err := parseAdminContestSchedule(request)
		if err != nil {
			writeJSONError(w, http.StatusBadRequest, err.Error())
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
		contestID, err := parseAdminContestID(r)
		if err != nil {
			writeJSONError(w, http.StatusBadRequest, "invalid contest id")
			return
		}
		request, err := decodeAdminContestUpsertRequest(r)
		if err != nil {
			writeJSONError(w, http.StatusBadRequest, "invalid json body")
			return
		}
		startsAt, endsAt, err := parseAdminContestSchedule(request)
		if err != nil {
			writeJSONError(w, http.StatusBadRequest, err.Error())
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
		contestID, err := parseAdminContestID(r)
		if err != nil {
			writeJSONError(w, http.StatusBadRequest, "invalid contest id")
			return
		}
		request, err := decodeAdminContestProblemsReplaceRequest(r)
		if err != nil {
			writeJSONError(w, http.StatusBadRequest, "invalid json body")
			return
		}
		problems, err := parseAdminContestProblemBindings(request)
		if err != nil {
			writeJSONError(w, http.StatusBadRequest, "invalid problem id")
			return
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
		contestID, err := parseAdminContestID(r)
		if err != nil {
			writeJSONError(w, http.StatusBadRequest, "invalid contest id")
			return
		}
		request, err := decodeAdminReasonOnlyRequest(r)
		if err != nil {
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
