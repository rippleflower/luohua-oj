package http

import (
	"crypto/rand"
	"encoding/base64"
	"net/http"
	"strings"
	"time"

	"github.com/example/oj3/apps/api/internal/auth"
	"github.com/example/oj3/apps/api/internal/contest"
	"github.com/example/oj3/apps/api/internal/contest_makeup"
	"github.com/example/oj3/apps/api/internal/problem"
	"github.com/example/oj3/apps/api/internal/submission"
	"github.com/google/uuid"
)

type authHandlerOptions struct {
	service           *auth.Service
	sessionCookieName string
	cookieSecure      bool
	sourceRoot        string
	redisAddr         string
	problemAdmin      problem.AdminManager
	contestAdmin      contest.AdminManager
	submissionAdmin   submission.AdminManager
	contestMakeup     contest_makeup.Reader
}

func registerHandler(opts authHandlerOptions) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		request, err := decodeAuthRegisterRequest(r)
		if err != nil {
			writeJSONError(w, http.StatusBadRequest, "invalid json body")
			return
		}
		if request.Password != request.ConfirmPassword {
			writeJSONError(w, http.StatusBadRequest, "password confirmation does not match")
			return
		}
		registered, err := opts.service.Register(r.Context(), auth.RegisterInput{
			Email:       request.Email,
			Username:    request.Username,
			Password:    request.Password,
			DisplayName: request.DisplayName,
		}, requestIP(r), r.UserAgent())
		if err != nil {
			writeJSONError(w, http.StatusBadRequest, err.Error())
			return
		}
		setSessionCookies(w, opts, registered.SessionToken, registered.ExpiresAt)
		writeJSON(w, http.StatusCreated, map[string]any{
			"user":      authUserResponse(registered.User),
			"expiresAt": registered.ExpiresAt.UTC().Format(time.RFC3339),
		})
	}
}

func loginHandler(opts authHandlerOptions) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		request, err := decodeAuthLoginRequest(r)
		if err != nil {
			writeJSONError(w, http.StatusBadRequest, "invalid json body")
			return
		}
		loggedIn, err := opts.service.Login(r.Context(), auth.LoginInput{
			Identifier: request.Identifier,
			Password:   request.Password,
			IP:         requestIP(r),
		}, r.UserAgent())
		if err != nil {
			writeJSONError(w, http.StatusBadRequest, err.Error())
			return
		}
		setSessionCookies(w, opts, loggedIn.SessionToken, loggedIn.ExpiresAt)
		writeJSON(w, http.StatusOK, map[string]any{
			"user":      authUserResponse(loggedIn.User),
			"expiresAt": loggedIn.ExpiresAt.UTC().Format(time.RFC3339),
		})
	}
}

func authMeHandler() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		user, ok := currentUser(r.Context())
		if !ok {
			writeJSONError(w, http.StatusUnauthorized, "authentication required")
			return
		}
		writeJSON(w, http.StatusOK, authUserResponse(user.User))
	}
}

func logoutHandler(opts authHandlerOptions) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		cookie, _ := r.Cookie(opts.sessionCookieName)
		if cookie != nil {
			_ = opts.service.Logout(r.Context(), cookie.Value)
		}
		clearSessionCookies(w, opts)
		writeJSON(w, http.StatusOK, map[string]bool{"ok": true})
	}
}

func changePasswordHandler(opts authHandlerOptions) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		user, ok := currentUser(r.Context())
		if !ok {
			writeJSONError(w, http.StatusUnauthorized, "authentication required")
			return
		}
		request, err := decodeAuthChangePasswordRequest(r)
		if err != nil {
			writeJSONError(w, http.StatusBadRequest, "invalid json body")
			return
		}
		if request.NewPassword != request.ConfirmPassword {
			writeJSONError(w, http.StatusBadRequest, "password confirmation does not match")
			return
		}
		if err := opts.service.ChangePassword(r.Context(), auth.ChangePasswordInput{
			UserID:           user.ID,
			CurrentPassword:  request.CurrentPassword,
			NewPassword:      request.NewPassword,
			SessionTokenHash: user.SessionTokenHash,
		}, requestIP(r)); err != nil {
			writeJSONError(w, http.StatusBadRequest, err.Error())
			return
		}
		writeJSON(w, http.StatusOK, map[string]bool{"ok": true})
	}
}

func listSessionsHandler(opts authHandlerOptions) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		user, ok := currentUser(r.Context())
		if !ok {
			writeJSONError(w, http.StatusUnauthorized, "authentication required")
			return
		}
		sessions, err := opts.service.ListSessions(r.Context(), user)
		if err != nil {
			writeJSONError(w, http.StatusInternalServerError, "session list failed")
			return
		}
		writeJSON(w, http.StatusOK, sessionsResponse(sessions))
	}
}

func revokeSessionHandler(opts authHandlerOptions) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		user, ok := currentUser(r.Context())
		if !ok {
			writeJSONError(w, http.StatusUnauthorized, "authentication required")
			return
		}
		request, err := decodeAuthRevokeSessionRequest(r)
		if err != nil {
			writeJSONError(w, http.StatusBadRequest, "invalid json body")
			return
		}
		sessionID, err := uuid.Parse(request.SessionID)
		if err != nil {
			writeJSONError(w, http.StatusBadRequest, "invalid sessionId")
			return
		}
		if err := opts.service.RevokeSession(r.Context(), user, sessionID); err != nil {
			writeJSONError(w, http.StatusBadRequest, err.Error())
			return
		}
		writeJSON(w, http.StatusOK, map[string]bool{"ok": true})
	}
}

func meSummaryHandler(opts authHandlerOptions) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		user, _ := currentUser(r.Context())
		summary, err := opts.service.MeSummary(r.Context(), user)
		if err != nil {
			writeJSONError(w, http.StatusInternalServerError, "failed to load summary")
			return
		}
		recentSubmissions := make([]map[string]any, 0, len(summary.RecentSubmissions))
		for _, item := range summary.RecentSubmissions {
			recentSubmissions = append(recentSubmissions, map[string]any{
				"id":        item.ID.String(),
				"status":    item.Status,
				"createdAt": item.CreatedAt.UTC().Format(time.RFC3339),
				"problem": map[string]any{
					"id":    item.ProblemID.String(),
					"slug":  item.ProblemSlug,
					"title": item.ProblemTitle,
				},
			})
		}
		contests := make([]map[string]any, 0, len(summary.Contests))
		for _, item := range summary.Contests {
			contests = append(contests, map[string]any{
				"id":       item.ID.String(),
				"slug":     item.Slug,
				"title":    item.Title,
				"status":   item.Status,
				"startsAt": item.StartsAt.UTC().Format(time.RFC3339),
				"endsAt":   item.EndsAt.UTC().Format(time.RFC3339),
			})
		}
		writeJSON(w, http.StatusOK, map[string]any{
			"user": authUserResponse(summary.User),
			"stats": map[string]any{
				"solvedCount":     summary.SolvedCount,
				"submissionCount": summary.SubmissionCount,
				"acceptedCount":   summary.AcceptedCount,
				"lastActiveAt":    formatNullableTime(summary.LastActiveAt),
			},
			"recentSubmissions": recentSubmissions,
			"contests":          contests,
		})
	}
}

func meSettingsHandler(opts authHandlerOptions) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		user, _ := currentUser(r.Context())
		settings, err := opts.service.MeSettings(r.Context(), user)
		if err != nil {
			writeJSONError(w, http.StatusInternalServerError, "failed to load settings")
			return
		}
		writeJSON(w, http.StatusOK, meSettingsResponse(settings))
	}
}

func updateProfileHandler(opts authHandlerOptions) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		user, _ := currentUser(r.Context())
		request, err := decodeAuthUpdateProfileRequest(r)
		if err != nil {
			writeJSONError(w, http.StatusBadRequest, "invalid json body")
			return
		}
		settings, err := opts.service.UpdateProfile(r.Context(), auth.UpdateProfileInput{
			UserID:      user.ID,
			DisplayName: strings.TrimSpace(request.DisplayName),
			Bio:         request.Bio,
			AvatarURL:   strings.TrimSpace(request.AvatarURL),
			IP:          requestIP(r),
		})
		if err != nil {
			writeJSONError(w, http.StatusBadRequest, err.Error())
			return
		}
		writeJSON(w, http.StatusOK, meSettingsResponse(settings))
	}
}

func updatePreferencesHandler(opts authHandlerOptions) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		user, _ := currentUser(r.Context())
		request, err := decodeAuthUpdatePreferencesRequest(r)
		if err != nil {
			writeJSONError(w, http.StatusBadRequest, "invalid json body")
			return
		}
		settings, err := opts.service.UpdatePreferences(r.Context(), auth.UpdatePreferencesInput{
			UserID:            user.ID,
			PreferredLocale:   strings.TrimSpace(request.PreferredLocale),
			PreferredLanguage: strings.TrimSpace(request.PreferredLanguage),
			IP:                requestIP(r),
		})
		if err != nil {
			writeJSONError(w, http.StatusBadRequest, err.Error())
			return
		}
		writeJSON(w, http.StatusOK, meSettingsResponse(settings))
	}
}

func setSessionCookies(w http.ResponseWriter, opts authHandlerOptions, sessionToken string, expiresAt time.Time) {
	http.SetCookie(w, &http.Cookie{
		Name:     opts.sessionCookieName,
		Value:    sessionToken,
		Path:     "/",
		Expires:  expiresAt,
		HttpOnly: true,
		Secure:   opts.cookieSecure,
		SameSite: http.SameSiteLaxMode,
	})
	http.SetCookie(w, &http.Cookie{
		Name:     "oj_csrf",
		Value:    randomCookieValue(),
		Path:     "/",
		Expires:  expiresAt,
		HttpOnly: false,
		Secure:   opts.cookieSecure,
		SameSite: http.SameSiteLaxMode,
	})
}

func clearSessionCookies(w http.ResponseWriter, opts authHandlerOptions) {
	http.SetCookie(w, &http.Cookie{Name: opts.sessionCookieName, Value: "", Path: "/", MaxAge: -1, HttpOnly: true, Secure: opts.cookieSecure, SameSite: http.SameSiteLaxMode})
	http.SetCookie(w, &http.Cookie{Name: "oj_csrf", Value: "", Path: "/", MaxAge: -1, HttpOnly: false, Secure: opts.cookieSecure, SameSite: http.SameSiteLaxMode})
}

func randomCookieValue() string {
	buf := make([]byte, 24)
	if _, err := rand.Read(buf); err != nil {
		return ""
	}
	return base64.RawURLEncoding.EncodeToString(buf)
}

func requestIP(r *http.Request) string {
	forwarded := strings.TrimSpace(r.Header.Get("X-Forwarded-For"))
	if forwarded != "" {
		return strings.Split(forwarded, ",")[0]
	}
	return r.RemoteAddr
}
