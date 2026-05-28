package http

import (
	"encoding/json"
	"net/http"
	"time"

	"github.com/example/oj3/apps/api/internal/auth"
	"github.com/example/oj3/apps/api/internal/contest"
	"github.com/example/oj3/apps/api/internal/problem"
	"github.com/example/oj3/apps/api/internal/submission"
	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"
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

func adminProblemsHandler(opts authHandlerOptions) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		items, err := opts.service.ListAdminProblems(r.Context())
		if err != nil {
			writeJSONError(w, http.StatusInternalServerError, "failed to load problems")
			return
		}
		response := make([]map[string]any, 0, len(items))
		for _, item := range items {
			response = append(response, adminProblemResponse(problem.AdminProblem{
				ID:               item.ID,
				ProblemNo:        item.ProblemNo,
				RouteCode:        opts.problemRouteCodec.Encode(item.ProblemNo),
				Slug:             item.Slug,
				Title:            item.Title,
				Difficulty:       item.Difficulty,
				TimeLimitMs:      item.TimeLimitMs,
				MemoryLimitKb:    item.MemoryLimitKb,
				Status:           item.Status,
				CurrentVersionNo: item.CurrentVersionNo,
				IsPublished:      item.IsPublished,
				SubmissionCount:  item.SubmissionCount,
				AcceptedRate:     item.AcceptedRate,
				UpdatedAt:        item.UpdatedAt,
			}))
		}
		writeJSON(w, http.StatusOK, response)
	}
}

func adminProblemCreateHandler(opts authHandlerOptions) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		actor, _ := currentUser(r.Context())
		var request struct {
			Slug          string `json:"slug"`
			Title         string `json:"title"`
			Difficulty    string `json:"difficulty"`
			TimeLimitMs   int    `json:"timeLimitMs"`
			MemoryLimitKb int    `json:"memoryLimitKb"`
			Reason        string `json:"reason"`
		}
		if err := json.NewDecoder(r.Body).Decode(&request); err != nil {
			writeJSONError(w, http.StatusBadRequest, "invalid json body")
			return
		}
		item, err := opts.problemAdmin.CreateAdmin(r.Context(), actor, problem.CreateAdminInput{
			Slug:          request.Slug,
			Title:         request.Title,
			Difficulty:    request.Difficulty,
			TimeLimitMs:   request.TimeLimitMs,
			MemoryLimitKb: request.MemoryLimitKb,
			Reason:        request.Reason,
			IP:            requestIP(r),
		})
		if err != nil {
			writeJSONError(w, http.StatusBadRequest, err.Error())
			return
		}
		writeJSON(w, http.StatusCreated, adminProblemResponse(item))
	}
}

func adminProblemUpdateHandler(opts authHandlerOptions) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		actor, _ := currentUser(r.Context())
		problemID, err := uuid.Parse(chi.URLParam(r, "problemID"))
		if err != nil {
			writeJSONError(w, http.StatusBadRequest, "invalid problem id")
			return
		}
		var request struct {
			Slug          string `json:"slug"`
			Title         string `json:"title"`
			Difficulty    string `json:"difficulty"`
			TimeLimitMs   int    `json:"timeLimitMs"`
			MemoryLimitKb int    `json:"memoryLimitKb"`
			Reason        string `json:"reason"`
		}
		if err := json.NewDecoder(r.Body).Decode(&request); err != nil {
			writeJSONError(w, http.StatusBadRequest, "invalid json body")
			return
		}
		item, err := opts.problemAdmin.UpdateAdmin(r.Context(), actor, problem.UpdateAdminInput{
			ProblemID:     problemID,
			Slug:          request.Slug,
			Title:         request.Title,
			Difficulty:    request.Difficulty,
			TimeLimitMs:   request.TimeLimitMs,
			MemoryLimitKb: request.MemoryLimitKb,
			Reason:        request.Reason,
			IP:            requestIP(r),
		})
		if err != nil {
			writeJSONError(w, http.StatusBadRequest, err.Error())
			return
		}
		writeJSON(w, http.StatusOK, adminProblemResponse(item))
	}
}

func adminProblemPublishHandler(opts authHandlerOptions) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		actor, _ := currentUser(r.Context())
		problemID, err := uuid.Parse(chi.URLParam(r, "problemID"))
		if err != nil {
			writeJSONError(w, http.StatusBadRequest, "invalid problem id")
			return
		}
		var request struct {
			Reason string `json:"reason"`
		}
		if err := json.NewDecoder(r.Body).Decode(&request); err != nil {
			writeJSONError(w, http.StatusBadRequest, "invalid json body")
			return
		}
		item, err := opts.problemAdmin.PublishAdmin(r.Context(), actor, problem.PublishAdminInput{
			ProblemID: problemID,
			Reason:    request.Reason,
			IP:        requestIP(r),
		})
		if err != nil {
			writeJSONError(w, http.StatusBadRequest, err.Error())
			return
		}
		writeJSON(w, http.StatusOK, adminProblemResponse(item))
	}
}

func adminContestsHandler(opts authHandlerOptions) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		items, err := opts.contestAdmin.ListAdmin(r.Context())
		if err != nil {
			writeJSONError(w, http.StatusInternalServerError, "failed to load contests")
			return
		}
		response := make([]map[string]any, 0, len(items))
		for _, item := range items {
			response = append(response, adminContestResponse(item))
		}
		writeJSON(w, http.StatusOK, response)
	}
}

func adminContestCreateHandler(opts authHandlerOptions) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		actor, _ := currentUser(r.Context())
		var request struct {
			Slug        string `json:"slug"`
			Title       string `json:"title"`
			Description string `json:"description"`
			Status      string `json:"status"`
			StartsAt    string `json:"startsAt"`
			EndsAt      string `json:"endsAt"`
			Reason      string `json:"reason"`
		}
		if err := json.NewDecoder(r.Body).Decode(&request); err != nil {
			writeJSONError(w, http.StatusBadRequest, "invalid json body")
			return
		}
		startsAt, err := time.Parse(time.RFC3339, request.StartsAt)
		if err != nil {
			writeJSONError(w, http.StatusBadRequest, "invalid startsAt")
			return
		}
		endsAt, err := time.Parse(time.RFC3339, request.EndsAt)
		if err != nil {
			writeJSONError(w, http.StatusBadRequest, "invalid endsAt")
			return
		}
		item, err := opts.contestAdmin.CreateAdmin(r.Context(), actor, contest.CreateAdminInput{
			Slug:        request.Slug,
			Title:       request.Title,
			Description: request.Description,
			Status:      request.Status,
			StartsAt:    startsAt,
			EndsAt:      endsAt,
			Reason:      request.Reason,
			IP:          requestIP(r),
		})
		if err != nil {
			writeJSONError(w, http.StatusBadRequest, err.Error())
			return
		}
		writeJSON(w, http.StatusCreated, adminContestResponse(item))
	}
}

func adminContestUpdateHandler(opts authHandlerOptions) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		actor, _ := currentUser(r.Context())
		contestID, err := uuid.Parse(chi.URLParam(r, "contestID"))
		if err != nil {
			writeJSONError(w, http.StatusBadRequest, "invalid contest id")
			return
		}
		var request struct {
			Slug        string `json:"slug"`
			Title       string `json:"title"`
			Description string `json:"description"`
			Status      string `json:"status"`
			StartsAt    string `json:"startsAt"`
			EndsAt      string `json:"endsAt"`
			Reason      string `json:"reason"`
		}
		if err := json.NewDecoder(r.Body).Decode(&request); err != nil {
			writeJSONError(w, http.StatusBadRequest, "invalid json body")
			return
		}
		startsAt, err := time.Parse(time.RFC3339, request.StartsAt)
		if err != nil {
			writeJSONError(w, http.StatusBadRequest, "invalid startsAt")
			return
		}
		endsAt, err := time.Parse(time.RFC3339, request.EndsAt)
		if err != nil {
			writeJSONError(w, http.StatusBadRequest, "invalid endsAt")
			return
		}
		item, err := opts.contestAdmin.UpdateAdmin(r.Context(), actor, contest.UpdateAdminInput{
			ContestID:   contestID,
			Slug:        request.Slug,
			Title:       request.Title,
			Description: request.Description,
			Status:      request.Status,
			StartsAt:    startsAt,
			EndsAt:      endsAt,
			Reason:      request.Reason,
			IP:          requestIP(r),
		})
		if err != nil {
			writeJSONError(w, http.StatusBadRequest, err.Error())
			return
		}
		writeJSON(w, http.StatusOK, adminContestResponse(item))
	}
}

func adminContestProblemsReplaceHandler(opts authHandlerOptions) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		actor, _ := currentUser(r.Context())
		contestID, err := uuid.Parse(chi.URLParam(r, "contestID"))
		if err != nil {
			writeJSONError(w, http.StatusBadRequest, "invalid contest id")
			return
		}
		var request struct {
			Problems []struct {
				ProblemID string `json:"problemId"`
				Code      string `json:"code"`
				Position  int    `json:"position"`
			} `json:"problems"`
			Reason string `json:"reason"`
		}
		if err := json.NewDecoder(r.Body).Decode(&request); err != nil {
			writeJSONError(w, http.StatusBadRequest, "invalid json body")
			return
		}
		problems := make([]contest.AdminProblemBindingInput, 0, len(request.Problems))
		for _, item := range request.Problems {
			problemID, err := uuid.Parse(item.ProblemID)
			if err != nil {
				writeJSONError(w, http.StatusBadRequest, "invalid problem id")
				return
			}
			problems = append(problems, contest.AdminProblemBindingInput{
				ProblemID: problemID,
				Code:      item.Code,
				Position:  item.Position,
			})
		}
		item, err := opts.contestAdmin.ReplaceProblemsAdmin(r.Context(), actor, contest.ReplaceProblemsInput{
			ContestID: contestID,
			Problems:  problems,
			Reason:    request.Reason,
			IP:        requestIP(r),
		})
		if err != nil {
			writeJSONError(w, http.StatusBadRequest, err.Error())
			return
		}
		writeJSON(w, http.StatusOK, adminContestResponse(item))
	}
}

func adminContestFreezeHandler(opts authHandlerOptions) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		actor, _ := currentUser(r.Context())
		contestID, err := uuid.Parse(chi.URLParam(r, "contestID"))
		if err != nil {
			writeJSONError(w, http.StatusBadRequest, "invalid contest id")
			return
		}
		var request struct {
			Reason string `json:"reason"`
		}
		if err := json.NewDecoder(r.Body).Decode(&request); err != nil {
			writeJSONError(w, http.StatusBadRequest, "invalid json body")
			return
		}
		item, err := opts.contestAdmin.FreezeAdmin(r.Context(), actor, contest.FreezeAdminInput{
			ContestID: contestID,
			Reason:    request.Reason,
			IP:        requestIP(r),
		})
		if err != nil {
			writeJSONError(w, http.StatusBadRequest, err.Error())
			return
		}
		writeJSON(w, http.StatusOK, adminContestResponse(item))
	}
}

func adminSubmissionsHandler(opts authHandlerOptions) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		items, err := opts.service.ListAdminSubmissions(r.Context())
		if err != nil {
			writeJSONError(w, http.StatusInternalServerError, "failed to load submissions")
			return
		}
		response := make([]map[string]any, 0, len(items))
		for _, item := range items {
			response = append(response, map[string]any{
				"id":           item.ID.String(),
				"username":     item.Username,
				"problemTitle": item.ProblemTitle,
				"language":     item.Language,
				"status":       item.Status,
				"createdAt":    item.CreatedAt.UTC().Format(time.RFC3339),
			})
		}
		writeJSON(w, http.StatusOK, response)
	}
}

func adminJudgeQueueHandler(opts authHandlerOptions) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		summary, err := opts.submissionAdmin.QueueSummary(r.Context())
		if err != nil {
			writeJSONError(w, http.StatusInternalServerError, "failed to load judge queue")
			return
		}
		writeJSON(w, http.StatusOK, judgeQueueSummaryResponse(summary))
	}
}

func adminSubmissionRejudgeHandler(opts authHandlerOptions) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		actor, _ := currentUser(r.Context())
		submissionID, err := uuid.Parse(chi.URLParam(r, "submissionID"))
		if err != nil {
			writeJSONError(w, http.StatusBadRequest, "invalid submission id")
			return
		}
		var request struct {
			Reason string `json:"reason"`
		}
		if err := json.NewDecoder(r.Body).Decode(&request); err != nil {
			writeJSONError(w, http.StatusBadRequest, "invalid json body")
			return
		}
		result, err := opts.submissionAdmin.RejudgeAdmin(r.Context(), actor, submission.RejudgeAdminInput{
			SubmissionID: submissionID,
			Reason:       request.Reason,
			IP:           requestIP(r),
		})
		if err != nil {
			writeJSONError(w, http.StatusBadRequest, err.Error())
			return
		}
		writeJSON(w, http.StatusOK, map[string]any{
			"submissionId":          result.SubmissionID.String(),
			"queue":                 result.Queue,
			"resultSnapshotVersion": result.ResultSnapshotVersion,
			"problemVersionId":      nullableUUIDStringPtr(result.ProblemVersionID),
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

func adminUserDetailResponse(item auth.AdminUserDetail) map[string]any {
	permissions := make([]string, 0, len(item.Permissions))
	for _, permission := range item.Permissions {
		permissions = append(permissions, string(permission))
	}
	return map[string]any{
		"id":          item.ID.String(),
		"email":       item.Email,
		"username":    item.Username,
		"role":        item.Role,
		"status":      item.Status,
		"displayName": item.DisplayName,
		"createdAt":   item.CreatedAt.UTC().Format(time.RFC3339),
		"updatedAt":   item.UpdatedAt.UTC().Format(time.RFC3339),
		"permissions": permissions,
		"profile": map[string]any{
			"displayName":       item.DisplayName,
			"bio":               item.Bio,
			"avatarUrl":         item.AvatarURL,
			"preferredLocale":   item.PreferredLocale,
			"preferredLanguage": item.PreferredLanguage,
		},
		"stats": map[string]any{
			"solvedCount":     item.SolvedCount,
			"submissionCount": item.SubmissionCount,
			"acceptedCount":   item.AcceptedCount,
			"lastActiveAt":    formatNullableTime(item.LastActiveAt),
		},
	}
}

func systemSettingsResponse(settings auth.SystemSettings) map[string]any {
	return map[string]any{
		"registrationEnabled": settings.RegistrationEnabled,
		"judgeQueuePaused":    settings.JudgeQueuePaused,
		"storageMode":         settings.StorageMode,
		"sourceRoot":          settings.SourceRoot,
		"redisAddr":           settings.RedisAddr,
		"updatedAt":           settings.UpdatedAt.UTC().Format(time.RFC3339),
	}
}

func announcementResponse(item auth.Announcement) map[string]any {
	return map[string]any{
		"id":        item.ID.String(),
		"title":     item.Title,
		"content":   item.Content,
		"status":    item.Status,
		"audience":  item.Audience,
		"createdAt": item.CreatedAt.UTC().Format(time.RFC3339),
		"updatedAt": item.UpdatedAt.UTC().Format(time.RFC3339),
	}
}

func adminContestResponse(item contest.AdminContest) map[string]any {
	problems := make([]map[string]any, 0, len(item.Problems))
	for _, problemItem := range item.Problems {
		problems = append(problems, map[string]any{
			"problemId":    problemItem.ProblemID.String(),
			"problemSlug":  problemItem.ProblemSlug,
			"problemTitle": problemItem.ProblemTitle,
			"code":         problemItem.Code,
			"position":     problemItem.Position,
		})
	}
	snapshots := make([]map[string]any, 0, len(item.Snapshots))
	for _, snapshot := range item.Snapshots {
		snapshots = append(snapshots, map[string]any{
			"id":           snapshot.ID.String(),
			"snapshotNo":   snapshot.SnapshotNo,
			"frozenAt":     snapshot.FrozenAt.UTC().Format(time.RFC3339),
			"problemCount": snapshot.ProblemCount,
		})
	}
	return map[string]any{
		"id":               item.ID.String(),
		"slug":             item.Slug,
		"title":            item.Title,
		"description":      item.Description,
		"status":           item.Status,
		"startsAt":         item.StartsAt.UTC().Format(time.RFC3339),
		"endsAt":           item.EndsAt.UTC().Format(time.RFC3339),
		"participantCount": item.ParticipantCount,
		"problemCount":     item.ProblemCount,
		"latestSnapshotNo": item.LatestSnapshotNo,
		"updatedAt":        item.UpdatedAt.UTC().Format(time.RFC3339),
		"problems":         problems,
		"snapshots":        snapshots,
	}
}

func judgeQueueSummaryResponse(summary submission.QueueSummary) map[string]any {
	recentTasks := make([]map[string]any, 0, len(summary.RecentTasks))
	for _, task := range summary.RecentTasks {
		recentTasks = append(recentTasks, map[string]any{
			"id":            task.ID,
			"type":          task.Type,
			"submissionId":  nullableUUIDStringPtr(task.SubmissionID),
			"state":         task.State,
			"createdAt":     formatNullableTime(task.CreatedAt),
			"nextProcessAt": formatNullableTime(task.NextProcessAt),
			"completedAt":   formatNullableTime(task.CompletedAt),
			"lastErr":       task.LastErr,
		})
	}
	return map[string]any{
		"queue":          summary.Queue,
		"paused":         summary.Paused,
		"latencySeconds": summary.LatencySeconds,
		"pending":        summary.Pending,
		"active":         summary.Active,
		"scheduled":      summary.Scheduled,
		"retry":          summary.Retry,
		"archived":       summary.Archived,
		"completed":      summary.Completed,
		"processedToday": summary.ProcessedToday,
		"failedToday":    summary.FailedToday,
		"recentTasks":    recentTasks,
		"updatedAt":      summary.UpdatedAt.UTC().Format(time.RFC3339),
	}
}

func adminProblemResponse(item problem.AdminProblem) map[string]any {
	return map[string]any{
		"id":               item.ID.String(),
		"problemNo":        item.ProblemNo,
		"routeCode":        item.RouteCode,
		"slug":             item.Slug,
		"title":            item.Title,
		"difficulty":       item.Difficulty,
		"timeLimitMs":      item.TimeLimitMs,
		"memoryLimitKb":    item.MemoryLimitKb,
		"status":           item.Status,
		"currentVersionNo": item.CurrentVersionNo,
		"isPublished":      item.IsPublished,
		"submissionCount":  item.SubmissionCount,
		"acceptedRate":     item.AcceptedRate,
		"updatedAt":        item.UpdatedAt.UTC().Format(time.RFC3339),
	}
}

func nullableUUIDStringPtr(value *uuid.UUID) any {
	if value == nil {
		return nil
	}
	return value.String()
}

func nullableUUIDString(value *uuid.UUID) any {
	if value == nil {
		return nil
	}
	return value.String()
}

func nullableString(value *string) any {
	if value == nil {
		return nil
	}
	return *value
}

func nullableRole(value *auth.Role) any {
	if value == nil {
		return nil
	}
	return *value
}
