package http

import (
	"encoding/json"
	"net/http"
	"time"

	"github.com/example/oj3/apps/api/internal/auth"
	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"
)

func adminUsersHandler(opts authHandlerOptions) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		items, err := opts.service.ListAdminUsers(r.Context())
		if err != nil {
			writeJSONError(w, http.StatusInternalServerError, "failed to load users")
			return
		}
		response := make([]map[string]any, 0, len(items))
		for _, item := range items {
			response = append(response, map[string]any{
				"id":          item.ID.String(),
				"email":       item.Email,
				"username":    item.Username,
				"role":        item.Role,
				"status":      item.Status,
				"displayName": item.DisplayName,
				"createdAt":   item.CreatedAt.UTC().Format(time.RFC3339),
				"updatedAt":   item.UpdatedAt.UTC().Format(time.RFC3339),
			})
		}
		writeJSON(w, http.StatusOK, response)
	}
}

func adminUserDetailHandler(opts authHandlerOptions) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		userID, err := uuid.Parse(chi.URLParam(r, "userID"))
		if err != nil {
			writeJSONError(w, http.StatusBadRequest, "invalid user id")
			return
		}
		item, err := opts.service.GetAdminUser(r.Context(), userID)
		if err != nil {
			writeJSONError(w, http.StatusNotFound, "user not found")
			return
		}
		writeJSON(w, http.StatusOK, adminUserDetailResponse(item))
	}
}

func adminUserPermissionsHandler(opts authHandlerOptions) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		userID, err := uuid.Parse(chi.URLParam(r, "userID"))
		if err != nil {
			writeJSONError(w, http.StatusBadRequest, "invalid user id")
			return
		}
		permissions, err := opts.service.GetAdminPermissions(r.Context(), userID)
		if err != nil {
			writeJSONError(w, http.StatusInternalServerError, "failed to load permissions")
			return
		}
		response := make([]string, 0, len(permissions))
		for _, permission := range permissions {
			response = append(response, string(permission))
		}
		writeJSON(w, http.StatusOK, response)
	}
}

func adminUserUpdateHandler(opts authHandlerOptions) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		actor, _ := currentUser(r.Context())
		userID, err := uuid.Parse(chi.URLParam(r, "userID"))
		if err != nil {
			writeJSONError(w, http.StatusBadRequest, "invalid user id")
			return
		}
		var request struct {
			Status      string `json:"status"`
			DisplayName string `json:"displayName"`
			Bio         string `json:"bio"`
			AvatarURL   string `json:"avatarUrl"`
			Reason      string `json:"reason"`
		}
		if err := json.NewDecoder(r.Body).Decode(&request); err != nil {
			writeJSONError(w, http.StatusBadRequest, "invalid json body")
			return
		}
		item, err := opts.service.UpdateAdminUser(r.Context(), auth.UpdateAdminUserInput{
			Actor:       actor,
			TargetID:    userID,
			Status:      request.Status,
			DisplayName: request.DisplayName,
			Bio:         request.Bio,
			AvatarURL:   request.AvatarURL,
			Reason:      request.Reason,
			IP:          requestIP(r),
		})
		if err != nil {
			writeJSONError(w, http.StatusBadRequest, err.Error())
			return
		}
		writeJSON(w, http.StatusOK, adminUserDetailResponse(item))
	}
}

func adminUserRoleHandler(opts authHandlerOptions) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		actor, _ := currentUser(r.Context())
		userID, err := uuid.Parse(chi.URLParam(r, "userID"))
		if err != nil {
			writeJSONError(w, http.StatusBadRequest, "invalid user id")
			return
		}
		var request struct {
			Role   string `json:"role"`
			Reason string `json:"reason"`
		}
		if err := json.NewDecoder(r.Body).Decode(&request); err != nil {
			writeJSONError(w, http.StatusBadRequest, "invalid json body")
			return
		}
		if err := opts.service.SetUserRole(r.Context(), actor, userID, auth.Role(request.Role), request.Reason, requestIP(r)); err != nil {
			writeJSONError(w, http.StatusBadRequest, err.Error())
			return
		}
		writeJSON(w, http.StatusOK, map[string]bool{"ok": true})
	}
}

func adminUserPermissionsUpdateHandler(opts authHandlerOptions) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		actor, _ := currentUser(r.Context())
		userID, err := uuid.Parse(chi.URLParam(r, "userID"))
		if err != nil {
			writeJSONError(w, http.StatusBadRequest, "invalid user id")
			return
		}
		var request struct {
			Permissions []string `json:"permissions"`
			Reason      string   `json:"reason"`
		}
		if err := json.NewDecoder(r.Body).Decode(&request); err != nil {
			writeJSONError(w, http.StatusBadRequest, "invalid json body")
			return
		}
		permissions := make([]auth.PermissionKey, 0, len(request.Permissions))
		for _, permission := range request.Permissions {
			permissions = append(permissions, auth.PermissionKey(permission))
		}
		if err := opts.service.ReplacePermissions(r.Context(), actor, userID, permissions, request.Reason, requestIP(r)); err != nil {
			writeJSONError(w, http.StatusBadRequest, err.Error())
			return
		}
		writeJSON(w, http.StatusOK, map[string]bool{"ok": true})
	}
}
