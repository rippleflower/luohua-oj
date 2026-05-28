package http

import (
	"time"

	"github.com/example/oj3/apps/api/internal/auth"
)

func authUserResponse(user auth.User) map[string]any {
	permissions := make([]string, 0, len(user.Permissions))
	for _, permission := range user.Permissions {
		permissions = append(permissions, string(permission))
	}
	return map[string]any{
		"id":          user.ID.String(),
		"email":       user.Email,
		"username":    user.Username,
		"role":        user.Role,
		"permissions": permissions,
		"displayName": user.DisplayName,
	}
}

func sessionsResponse(sessions []auth.Session) []map[string]any {
	response := make([]map[string]any, 0, len(sessions))
	for _, session := range sessions {
		response = append(response, map[string]any{
			"id":         session.ID.String(),
			"current":    session.Current,
			"ip":         session.IP,
			"userAgent":  session.UserAgent,
			"expiresAt":  session.ExpiresAt.UTC().Format(time.RFC3339),
			"lastSeenAt": formatNullableTime(session.LastSeenAt),
			"createdAt":  session.CreatedAt.UTC().Format(time.RFC3339),
		})
	}
	return response
}

func meSettingsResponse(settings auth.MeSettings) map[string]any {
	return map[string]any{
		"user": authUserResponse(settings.User),
		"profile": map[string]any{
			"email":       settings.Email,
			"username":    settings.Username,
			"displayName": settings.DisplayName,
			"bio":         settings.Bio,
			"avatarUrl":   settings.AvatarURL,
		},
		"preferences": map[string]any{
			"preferredLocale":   settings.PreferredLocale,
			"preferredLanguage": settings.PreferredLanguage,
		},
		"sessions": sessionsResponse(settings.Sessions),
	}
}

func formatNullableTime(value *time.Time) any {
	if value == nil {
		return nil
	}
	return value.UTC().Format(time.RFC3339)
}
