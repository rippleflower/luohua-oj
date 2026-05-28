package http

import (
	"time"

	"github.com/example/oj3/apps/api/internal/contest_makeup"
)

func mapContestMakeupItemResponse(item contest_makeup.Item) map[string]any {
	response := map[string]any{
		"problemId":       item.ProblemID.String(),
		"problemCode":     item.ProblemCode,
		"problemSlug":     item.ProblemSlug,
		"problemTitle":    item.ProblemTitle,
		"difficulty":      item.Difficulty,
		"category":        item.Category,
		"lastStatus":      item.LastStatus,
		"severityRank":    item.SeverityRank,
		"reasonSummary":   item.ReasonSummary,
		"suggestedAction": item.SuggestedAction,
	}
	if item.AttemptCount > 0 {
		response["attemptCount"] = item.AttemptCount
	}
	return response
}

func mapContestMakeupListResponse(list contest_makeup.List) map[string]any {
	responseItems := make([]map[string]any, 0, len(list.Items))
	for _, item := range list.Items {
		responseItems = append(responseItems, mapContestMakeupItemResponse(item))
	}
	return map[string]any{
		"contestSlug": list.ContestSlug,
		"generatedAt": list.GeneratedAt.UTC().Format(time.RFC3339),
		"items":       responseItems,
	}
}
