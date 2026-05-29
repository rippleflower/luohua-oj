package http

import (
	"time"

	"github.com/example/oj3/apps/api/internal/auth"
	"github.com/example/oj3/apps/api/internal/contest"
	"github.com/example/oj3/apps/api/internal/problem"
	"github.com/example/oj3/apps/api/internal/submission"
	"github.com/google/uuid"
)

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

func adminProblemDetailResponse(item problem.AdminProblemDetail) map[string]any {
	response := adminProblemResponse(item.AdminProblem)
	response["statementJson"] = item.StatementJSON
	response["samples"] = item.Samples
	response["tags"] = item.Tags
	return response
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
