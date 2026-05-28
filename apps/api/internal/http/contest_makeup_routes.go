package http

import "github.com/go-chi/chi/v5"

func mountContestMakeupRoutes(r chi.Router, opts authHandlerOptions) {
	if opts.contestMakeup == nil || opts.service == nil {
		return
	}
	r.With(requireAuthenticated(opts.service, opts.sessionCookieName)).
		Get("/contests/{slug}/makeup-list", contestMakeupListHandler(opts))
}
