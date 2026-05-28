package http

import (
	"encoding/json"
	"time"

	"github.com/example/oj3/apps/api/internal/problem"
)

type problemSummaryResponse struct {
	ID           string   `json:"id"`
	ProblemNo    int64    `json:"problemNo"`
	RouteCode    string   `json:"routeCode"`
	Slug         string   `json:"slug"`
	Title        string   `json:"title"`
	Difficulty   string   `json:"difficulty"`
	Tags         []string `json:"tags"`
	AcceptedRate float64  `json:"acceptedRate"`
}

type problemDetailResponse struct {
	ID            string          `json:"id"`
	ProblemNo     int64           `json:"problemNo"`
	RouteCode     string          `json:"routeCode"`
	Slug          string          `json:"slug"`
	Title         string          `json:"title"`
	Difficulty    string          `json:"difficulty"`
	StatementJSON json.RawMessage `json:"statementJson"`
	SamplesJSON   json.RawMessage `json:"samplesJson"`
	LimitsJSON    json.RawMessage `json:"limitsJson"`
	MetadataJSON  json.RawMessage `json:"metadataJson"`
	UpdatedAt     string          `json:"updatedAt"`
}

func mapProblemSummaryResponse(item problem.Summary) problemSummaryResponse {
	return problemSummaryResponse{
		ID:           item.ID.String(),
		ProblemNo:    item.ProblemNo,
		RouteCode:    item.RouteCode,
		Slug:         item.Slug,
		Title:        item.Title,
		Difficulty:   item.Difficulty,
		Tags:         item.Tags,
		AcceptedRate: item.AcceptedRate,
	}
}

func mapProblemDetailResponse(item problem.Detail) problemDetailResponse {
	return problemDetailResponse{
		ID:            item.ID.String(),
		ProblemNo:     item.ProblemNo,
		RouteCode:     item.RouteCode,
		Slug:          item.Slug,
		Title:         item.Title,
		Difficulty:    item.Difficulty,
		StatementJSON: item.StatementJSON,
		SamplesJSON:   item.SamplesJSON,
		LimitsJSON:    item.LimitsJSON,
		MetadataJSON:  item.MetadataJSON,
		UpdatedAt:     item.UpdatedAt.UTC().Format(time.RFC3339),
	}
}
