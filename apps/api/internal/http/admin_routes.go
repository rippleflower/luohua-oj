package http

import (
	"github.com/example/oj3/apps/api/internal/auth"
	"github.com/go-chi/chi/v5"
)

func mountAdminRoutes(r chi.Router, opts authHandlerOptions) {
	r.Route("/admin", func(r chi.Router) {
		r.Use(requireAuthenticated(opts.service, opts.sessionCookieName))

		r.With(requireAdminPermission(auth.PermissionDashboardView)).Get("/me", adminDashboardHandler(opts))
		r.With(requireAdminPermission(auth.PermissionUsersView)).Get("/users", adminUsersHandler(opts))
		r.With(requireAdminPermission(auth.PermissionUsersView)).Get("/users/{userID}", adminUserDetailHandler(opts))
		r.With(requireAdminPermission(auth.PermissionUsersView)).Get("/users/{userID}/permissions", adminUserPermissionsHandler(opts))
		r.With(requireAdminPermission(auth.PermissionUsersEdit), requireCSRF).Patch("/users/{userID}", adminUserUpdateHandler(opts))
		r.With(requireAdminPermission(auth.PermissionUsersRoles), requireCSRF).Post("/users/{userID}/role", adminUserRoleHandler(opts))
		r.With(requireAdminPermission(auth.PermissionUsersRoles), requireCSRF).Put("/users/{userID}/permissions", adminUserPermissionsUpdateHandler(opts))

		r.With(requireAdminPermission(auth.PermissionProblemsView)).Get("/problems", adminProblemsHandler(opts))
		r.With(requireAdminPermission(auth.PermissionProblemsEdit), requireCSRF).Post("/problems", adminProblemCreateHandler(opts))
		r.With(requireAdminPermission(auth.PermissionProblemsEdit), requireCSRF).Patch("/problems/{problemID}", adminProblemUpdateHandler(opts))
		r.With(requireAdminPermission(auth.PermissionProblemsPublish), requireCSRF).Post("/problems/{problemID}/publish", adminProblemPublishHandler(opts))
		r.With(requireAdminPermission(auth.PermissionContestsView)).Get("/contests", adminContestsHandler(opts))
		r.With(requireAdminPermission(auth.PermissionContestsEdit), requireCSRF).Post("/contests", adminContestCreateHandler(opts))
		r.With(requireAdminPermission(auth.PermissionContestsEdit), requireCSRF).Patch("/contests/{contestID}", adminContestUpdateHandler(opts))
		r.With(requireAdminPermission(auth.PermissionContestsEdit), requireCSRF).Put("/contests/{contestID}/problems", adminContestProblemsReplaceHandler(opts))
		r.With(requireAdminPermission(auth.PermissionContestsPublish), requireCSRF).Post("/contests/{contestID}/freeze", adminContestFreezeHandler(opts))
		r.With(requireAdminPermission(auth.PermissionSubmissionsView)).Get("/submissions", adminSubmissionsHandler(opts))
		r.With(requireAdminPermission(auth.PermissionSubmissionsView)).Get("/judge/queue", adminJudgeQueueHandler(opts))
		r.With(requireAdminPermission(auth.PermissionSubmissionsRedo), requireCSRF).Post("/submissions/{submissionID}/rejudge", adminSubmissionRejudgeHandler(opts))
		r.With(requireAdminPermission(auth.PermissionAuditView)).Get("/audit", adminAuditHandler(opts))

		r.With(requireAdminPermission(auth.PermissionSystemView)).Get("/system/settings", adminSystemSettingsHandler(opts))
		r.With(requireAdminPermission(auth.PermissionSystemEdit), requireCSRF).Patch("/system/settings", adminSystemSettingsUpdateHandler(opts))

		r.With(requireAdminPermission(auth.PermissionAnnView)).Get("/announcements", adminAnnouncementsHandler(opts))
		r.With(requireAdminPermission(auth.PermissionAnnEdit), requireCSRF).Post("/announcements", adminAnnouncementCreateHandler(opts))
		r.With(requireAdminPermission(auth.PermissionAnnEdit), requireCSRF).Patch("/announcements/{announcementID}", adminAnnouncementUpdateHandler(opts))
	})
}
