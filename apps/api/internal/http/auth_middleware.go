package http

import (
	"context"
	"net/http"

	"github.com/example/oj3/apps/api/internal/auth"
)

type currentUserContextKey struct{}

func requireAuthenticated(authService *auth.Service, sessionCookieName string) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			cookie, err := r.Cookie(sessionCookieName)
			if err != nil || cookie.Value == "" {
				writeJSONError(w, http.StatusUnauthorized, "authentication required")
				return
			}
			user, err := authService.Authenticate(r.Context(), cookie.Value)
			if err != nil {
				writeJSONError(w, http.StatusUnauthorized, "invalid session")
				return
			}
			ctx := context.WithValue(r.Context(), currentUserContextKey{}, user)
			next.ServeHTTP(w, r.WithContext(ctx))
		})
	}
}

func requireAdminPermission(permission auth.PermissionKey) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			user, ok := currentUser(r.Context())
			if !ok {
				writeJSONError(w, http.StatusUnauthorized, "authentication required")
				return
			}
			if user.Role == auth.RoleUser || !user.HasPermission(permission) {
				writeJSONError(w, http.StatusForbidden, "forbidden")
				return
			}
			next.ServeHTTP(w, r)
		})
	}
}

func requireCSRF(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		cookie, err := r.Cookie("oj_csrf")
		if err != nil || cookie.Value == "" || r.Header.Get("X-CSRF-Token") == "" || cookie.Value != r.Header.Get("X-CSRF-Token") {
			writeJSONError(w, http.StatusForbidden, "csrf token required")
			return
		}
		next.ServeHTTP(w, r)
	})
}

func currentUser(ctx context.Context) (auth.AuthenticatedUser, bool) {
	user, ok := ctx.Value(currentUserContextKey{}).(auth.AuthenticatedUser)
	return user, ok
}
