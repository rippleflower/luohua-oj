package http

import (
	"encoding/json"
	"net/http"
	"time"

	"github.com/example/oj3/apps/api/internal/auth"
)

func adminSystemSettingsHandler(opts authHandlerOptions) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		settings, err := opts.service.GetSystemSettings(r.Context(), opts.sourceRoot, opts.redisAddr)
		if err != nil {
			writeJSONError(w, http.StatusInternalServerError, "failed to load system settings")
			return
		}
		writeJSON(w, http.StatusOK, systemSettingsResponse(settings))
	}
}

func adminSystemSettingsUpdateHandler(opts authHandlerOptions) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		actor, _ := currentUser(r.Context())
		var request struct {
			RegistrationEnabled bool   `json:"registrationEnabled"`
			JudgeQueuePaused    bool   `json:"judgeQueuePaused"`
			StorageMode         string `json:"storageMode"`
		}
		if err := json.NewDecoder(r.Body).Decode(&request); err != nil {
			writeJSONError(w, http.StatusBadRequest, "invalid json body")
			return
		}
		settings, err := opts.service.UpdateSystemSettings(r.Context(), actor, auth.SystemSettings{
			RegistrationEnabled: request.RegistrationEnabled,
			JudgeQueuePaused:    request.JudgeQueuePaused,
			StorageMode:         request.StorageMode,
			SourceRoot:          opts.sourceRoot,
			RedisAddr:           opts.redisAddr,
			UpdatedAt:           time.Now(),
		}, requestIP(r))
		if err != nil {
			writeJSONError(w, http.StatusBadRequest, err.Error())
			return
		}
		writeJSON(w, http.StatusOK, systemSettingsResponse(settings))
	}
}
