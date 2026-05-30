package http

import "github.com/go-chi/chi/v5"

func mountAuthRoutes(r chi.Router, opts authHandlerOptions) {
	r.Route("/auth", func(r chi.Router) {
		r.Post("/register", registerHandler(opts))
		r.Post("/login", loginHandler(opts))
		r.With(requireAuthenticated(opts.service, opts.sessionCookieName)).Get("/me", authMeHandler())
		r.With(requireAuthenticated(opts.service, opts.sessionCookieName), requireCSRF).Post("/logout", logoutHandler(opts))
		r.With(requireAuthenticated(opts.service, opts.sessionCookieName), requireCSRF).Post("/password/change", changePasswordHandler(opts))
		r.With(requireAuthenticated(opts.service, opts.sessionCookieName)).Get("/sessions", listSessionsHandler(opts))
		r.With(requireAuthenticated(opts.service, opts.sessionCookieName), requireCSRF).Post("/sessions/revoke", revokeSessionHandler(opts))
	})

	r.Route("/me", func(r chi.Router) {
		r.Use(requireAuthenticated(opts.service, opts.sessionCookieName))
		r.Get("/summary", meSummaryHandler(opts))
		r.Get("/settings", meSettingsHandler(opts))
		r.With(requireCSRF).Patch("/profile", updateProfileHandler(opts))
		r.With(requireCSRF).Patch("/preferences", updatePreferencesHandler(opts))
	})
}
