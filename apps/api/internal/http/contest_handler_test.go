package http_test

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/example/oj3/apps/api/internal/contest"
	apphttp "github.com/example/oj3/apps/api/internal/http"
	"github.com/google/uuid"
	"github.com/stretchr/testify/require"
)

type fakeContestReader struct {
	list   []contest.Summary
	detail contest.Detail
}

func (f *fakeContestReader) List(ctx context.Context) ([]contest.Summary, error) {
	return f.list, nil
}

func (f *fakeContestReader) GetBySlug(ctx context.Context, slug string) (contest.Detail, error) {
	return f.detail, nil
}

func TestListContests(t *testing.T) {
	reader := &fakeContestReader{
		list: []contest.Summary{
			{
				ID:               uuid.New(),
				Slug:             "spring-open",
				Title:            "Spring Open",
				Status:           "RUNNING",
				StartsAt:         time.Date(2026, 5, 16, 11, 0, 0, 0, time.UTC),
				EndsAt:           time.Date(2026, 5, 16, 13, 0, 0, 0, time.UTC),
				DurationLabel:    "2 hours",
				ProblemCount:     6,
				ParticipantCount: 100,
				Blurb:            "fast contest",
			},
		},
	}

	req := httptest.NewRequest(http.MethodGet, "/contests", nil)
	rec := httptest.NewRecorder()

	apphttp.NewRouter(apphttp.RouterOptions{ContestReader: reader}).ServeHTTP(rec, req)

	require.Equal(t, http.StatusOK, rec.Code)
	require.Contains(t, rec.Body.String(), `"slug":"spring-open"`)
	require.Contains(t, rec.Body.String(), `"duration":"2 hours"`)
}

func TestGetContest(t *testing.T) {
	contestID := uuid.New()
	reader := &fakeContestReader{
		detail: contest.Detail{
			Summary: contest.Summary{
				ID:               contestID,
				Slug:             "spring-open",
				Title:            "Spring Open",
				Status:           "RUNNING",
				StartsAt:         time.Date(2026, 5, 16, 11, 0, 0, 0, time.UTC),
				EndsAt:           time.Date(2026, 5, 16, 13, 0, 0, 0, time.UTC),
				DurationLabel:    "2 hours",
				ProblemCount:     6,
				ParticipantCount: 100,
				Blurb:            "fast contest",
			},
			RankSummary: "Top 10%",
			Remaining:   "00:30:00",
			RecentSubmissions: []contest.RecentSubmission{
				{ID: "sub-1", ProblemCode: "A", Status: "ACCEPTED", At: "1 minute ago"},
			},
			Problems: []contest.ProblemSnapshot{
				{Code: "A", Slug: "two-sum", Title: "Two Sum", Difficulty: "EASY", Status: "SOLVED"},
			},
			UpdatedAt: time.Date(2026, 5, 16, 11, 30, 0, 0, time.UTC),
		},
	}

	req := httptest.NewRequest(http.MethodGet, "/contests/spring-open", nil)
	rec := httptest.NewRecorder()

	apphttp.NewRouter(apphttp.RouterOptions{ContestReader: reader}).ServeHTTP(rec, req)

	require.Equal(t, http.StatusOK, rec.Code)
	require.Contains(t, rec.Body.String(), `"id":"`+contestID.String()+`"`)
	require.Contains(t, rec.Body.String(), `"rankSummary":"Top 10%"`)
	require.Contains(t, rec.Body.String(), `"slug":"two-sum"`)
	require.Contains(t, rec.Body.String(), `"recentSubmissions":[{"id":"sub-1","problemCode":"A","status":"ACCEPTED","at":"1 minute ago"}]`)
}

func TestGetContestWithLegacyProblemSnapshot(t *testing.T) {
	reader := &fakeContestReader{
		detail: contest.Detail{
			Summary: contest.Summary{
				ID:               uuid.New(),
				Slug:             "legacy-open",
				Title:            "Legacy Open",
				Status:           "ENDED",
				StartsAt:         time.Date(2026, 5, 1, 11, 0, 0, 0, time.UTC),
				EndsAt:           time.Date(2026, 5, 1, 13, 0, 0, 0, time.UTC),
				DurationLabel:    "2 hours",
				ProblemCount:     1,
				ParticipantCount: 16,
				Blurb:            "legacy",
			},
			RankSummary:       "Completed",
			Remaining:         "比赛已结束",
			RecentSubmissions: nil,
			Problems: []contest.ProblemSnapshot{
				{Code: "A", Title: "Legacy Problem", Difficulty: "EASY", Status: "LOCKED"},
			},
			UpdatedAt: time.Date(2026, 5, 1, 13, 0, 0, 0, time.UTC),
		},
	}

	req := httptest.NewRequest(http.MethodGet, "/contests/legacy-open", nil)
	rec := httptest.NewRecorder()

	apphttp.NewRouter(apphttp.RouterOptions{ContestReader: reader}).ServeHTTP(rec, req)

	require.Equal(t, http.StatusOK, rec.Code)
	require.Contains(t, rec.Body.String(), `"title":"Legacy Problem"`)
}
