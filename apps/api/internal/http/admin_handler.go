package http

import (
	"net/http"
	"time"
)

func adminDashboardHandler(opts authHandlerOptions) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		actor, _ := currentUser(r.Context())
		dashboard, err := opts.service.AdminDashboard(r.Context(), actor)
		if err != nil {
			writeJSONError(w, http.StatusInternalServerError, "failed to load dashboard")
			return
		}
		alerts := make([]map[string]string, 0, len(dashboard.Alerts))
		for _, alert := range dashboard.Alerts {
			alerts = append(alerts, map[string]string{
				"title":       alert.Title,
				"tone":        alert.Tone,
				"description": alert.Description,
			})
		}
		writeJSON(w, http.StatusOK, map[string]any{
			"actor": authUserResponse(dashboard.Actor),
			"metrics": map[string]any{
				"totalUsers":         dashboard.TotalUsers,
				"activeProblems":     dashboard.ActiveProblems,
				"runningContests":    dashboard.RunningContests,
				"submissions24h":     dashboard.Submissions24h,
				"pendingSubmissions": dashboard.PendingSubmissions,
			},
			"alerts": alerts,
		})
	}
}

func adminAuditHandler(opts authHandlerOptions) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		items, err := opts.service.ListAuditEvents(r.Context())
		if err != nil {
			writeJSONError(w, http.StatusInternalServerError, "failed to load audit")
			return
		}
		response := make([]map[string]any, 0, len(items))
		for _, item := range items {
			response = append(response, map[string]any{
				"id": item.ID.String(),
				"actor": map[string]any{
					"id":       nullableUUIDString(item.ActorUserID),
					"username": nullableString(item.ActorUsername),
					"role":     nullableRole(item.ActorRole),
				},
				"action":     item.Action,
				"targetType": item.TargetType,
				"targetId":   item.TargetID,
				"diff":       item.Diff,
				"reason":     item.Reason,
				"ip":         item.IP,
				"createdAt":  item.CreatedAt.UTC().Format(time.RFC3339),
			})
		}
		writeJSON(w, http.StatusOK, response)
	}
}
