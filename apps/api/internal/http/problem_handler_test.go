package http_test

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	apphttp "github.com/example/oj3/apps/api/internal/http"
	"github.com/example/oj3/apps/api/internal/problem"
	"github.com/google/uuid"
	"github.com/stretchr/testify/require"
)

type fakeProblemReader struct {
	list   []problem.Summary
	detail problem.Detail
}

func (f *fakeProblemReader) List(ctx context.Context) ([]problem.Summary, error) {
	return f.list, nil
}

func (f *fakeProblemReader) GetBySlug(ctx context.Context, slug string) (problem.Detail, error) {
	return f.detail, nil
}

func (f *fakeProblemReader) GetByRouteCode(ctx context.Context, routeCode string) (problem.Detail, error) {
	return f.detail, nil
}

func TestListProblems(t *testing.T) {
	reader := &fakeProblemReader{
		list: []problem.Summary{
			{
				ID:           uuid.New(),
				ProblemNo:    12,
				RouteCode:    "ABC123",
				Slug:         "two-sum",
				Title:        "Two Sum",
				Difficulty:   "EASY",
				Tags:         []string{"array", "hash-table"},
				AcceptedRate: 62.4,
				UpdatedAt:    time.Now().UTC(),
			},
		},
	}

	req := httptest.NewRequest(http.MethodGet, "/problems", nil)
	rec := httptest.NewRecorder()

	apphttp.NewRouter(apphttp.RouterOptions{ProblemReader: reader}).ServeHTTP(rec, req)

	require.Equal(t, http.StatusOK, rec.Code)
	require.Contains(t, rec.Body.String(), `"problemNo":12`)
	require.Contains(t, rec.Body.String(), `"routeCode":"ABC123"`)
	require.Contains(t, rec.Body.String(), `"slug":"two-sum"`)
	require.Contains(t, rec.Body.String(), `"acceptedRate":62.4`)
	require.Contains(t, rec.Body.String(), `"tags":["array","hash-table"]`)
}

func TestGetProblem(t *testing.T) {
	problemID := uuid.New()
	reader := &fakeProblemReader{
		detail: problem.Detail{
			ID:            problemID,
			ProblemNo:     12,
			RouteCode:     "ABC123",
			Slug:          "two-sum",
			Title:         "Two Sum",
			Difficulty:    "EASY",
			StatementJSON: []byte(`[{"kind":"markdown","section":"statement","content":"desc"}]`),
			SamplesJSON:   []byte(`[]`),
			LimitsJSON:    []byte(`{"timeLimitMs":1000,"memoryLimitKb":262144}`),
			MetadataJSON:  []byte(`{"published":true}`),
			UpdatedAt:     time.Date(2026, 5, 16, 10, 0, 0, 0, time.UTC),
		},
	}

	req := httptest.NewRequest(http.MethodGet, "/problems/two-sum", nil)
	rec := httptest.NewRecorder()

	apphttp.NewRouter(apphttp.RouterOptions{ProblemReader: reader}).ServeHTTP(rec, req)

	require.Equal(t, http.StatusOK, rec.Code)
	require.Contains(t, rec.Body.String(), `"id":"`+problemID.String()+`"`)
	require.Contains(t, rec.Body.String(), `"routeCode":"ABC123"`)
	require.Contains(t, rec.Body.String(), `"statementJson":[{"kind":"markdown","section":"statement","content":"desc"}]`)
	require.Contains(t, rec.Body.String(), `"updatedAt":"2026-05-16T10:00:00Z"`)
}

func TestGetProblemByRouteCode(t *testing.T) {
	problemID := uuid.New()
	reader := &fakeProblemReader{
		detail: problem.Detail{
			ID:            problemID,
			ProblemNo:     12,
			RouteCode:     "ABC123",
			Slug:          "two-sum",
			Title:         "Two Sum",
			Difficulty:    "EASY",
			StatementJSON: []byte(`[{"kind":"markdown","section":"statement","content":"desc"}]`),
			SamplesJSON:   []byte(`[]`),
			LimitsJSON:    []byte(`{"timeLimitMs":1000,"memoryLimitKb":262144}`),
			MetadataJSON:  []byte(`{"published":true}`),
			UpdatedAt:     time.Date(2026, 5, 16, 10, 0, 0, 0, time.UTC),
		},
	}

	req := httptest.NewRequest(http.MethodGet, "/problems/code/ABC123", nil)
	rec := httptest.NewRecorder()

	apphttp.NewRouter(apphttp.RouterOptions{ProblemReader: reader}).ServeHTTP(rec, req)

	require.Equal(t, http.StatusOK, rec.Code)
	require.Contains(t, rec.Body.String(), `"routeCode":"ABC123"`)
	require.Contains(t, rec.Body.String(), `"problemNo":12`)
}
