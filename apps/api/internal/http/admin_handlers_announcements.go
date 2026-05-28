package http

import (
	"encoding/json"
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"
)

func adminAnnouncementsHandler(opts authHandlerOptions) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		items, err := opts.service.ListAnnouncements(r.Context())
		if err != nil {
			writeJSONError(w, http.StatusInternalServerError, "failed to load announcements")
			return
		}
		response := make([]map[string]any, 0, len(items))
		for _, item := range items {
			response = append(response, announcementResponse(item))
		}
		writeJSON(w, http.StatusOK, response)
	}
}

func adminAnnouncementCreateHandler(opts authHandlerOptions) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		actor, _ := currentUser(r.Context())
		var request struct {
			Title    string `json:"title"`
			Content  string `json:"content"`
			Status   string `json:"status"`
			Audience string `json:"audience"`
		}
		if err := json.NewDecoder(r.Body).Decode(&request); err != nil {
			writeJSONError(w, http.StatusBadRequest, "invalid json body")
			return
		}
		item, err := opts.service.CreateAnnouncement(r.Context(), actor, request.Title, request.Content, request.Status, request.Audience, requestIP(r))
		if err != nil {
			writeJSONError(w, http.StatusBadRequest, err.Error())
			return
		}
		writeJSON(w, http.StatusCreated, announcementResponse(item))
	}
}

func adminAnnouncementUpdateHandler(opts authHandlerOptions) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		actor, _ := currentUser(r.Context())
		announcementID, err := uuid.Parse(chi.URLParam(r, "announcementID"))
		if err != nil {
			writeJSONError(w, http.StatusBadRequest, "invalid announcement id")
			return
		}
		var request struct {
			Title    string `json:"title"`
			Content  string `json:"content"`
			Status   string `json:"status"`
			Audience string `json:"audience"`
		}
		if err := json.NewDecoder(r.Body).Decode(&request); err != nil {
			writeJSONError(w, http.StatusBadRequest, "invalid json body")
			return
		}
		item, err := opts.service.UpdateAnnouncement(r.Context(), actor, announcementID, request.Title, request.Content, request.Status, request.Audience, requestIP(r))
		if err != nil {
			writeJSONError(w, http.StatusBadRequest, err.Error())
			return
		}
		writeJSON(w, http.StatusOK, announcementResponse(item))
	}
}
