package http

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/example/oj3/apps/api/internal/auth"
	"github.com/example/oj3/apps/api/internal/contest_makeup"
	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"
	"github.com/stretchr/testify/require"
)

func TestContestMakeupListHandlerReturnsConflictWhenContestNotEnded(t *testing.T) {
	handler := contestMakeupListHandler(authHandlerOptions{
		contestMakeup: fakeContestMakeupReader{err: contest_makeup.ErrContestNotEnded},
	})

	req := httptest.NewRequest(http.MethodGet, "/contests/spring-open/makeup-list", nil)
	req = withContestSlug(req, "spring-open")
	req = withAuthenticatedUser(req)
	rec := httptest.NewRecorder()

	handler.ServeHTTP(rec, req)

	require.Equal(t, http.StatusConflict, rec.Code)
	require.Contains(t, rec.Body.String(), "contest is not ended")
}

func TestContestMakeupListHandlerReturnsList(t *testing.T) {
	handler := contestMakeupListHandler(authHandlerOptions{
		contestMakeup: fakeContestMakeupReader{
			list: contest_makeup.List{
				ContestSlug: "april-grand-prix",
				GeneratedAt: time.Date(2026, 5, 26, 9, 0, 0, 0, time.UTC),
				Items: []contest_makeup.Item{
					{
						ProblemID:       uuid.New(),
						ProblemCode:     "D",
						ProblemSlug:     "problem-d",
						ProblemTitle:    "Problem D",
						Difficulty:      "MEDIUM",
						Category:        "ATTEMPTED_UNSOLVED",
						LastStatus:      "WRONG_ANSWER",
						AttemptCount:    2,
						SeverityRank:    3,
						ReasonSummary:   "最近一次结果是 WA",
						SuggestedAction: "重读题意并补样例",
					},
				},
			},
		},
	})

	req := httptest.NewRequest(http.MethodGet, "/contests/april-grand-prix/makeup-list", nil)
	req = withContestSlug(req, "april-grand-prix")
	req = withAuthenticatedUser(req)
	rec := httptest.NewRecorder()

	handler.ServeHTTP(rec, req)

	require.Equal(t, http.StatusOK, rec.Code)
	require.Contains(t, rec.Body.String(), `"contestSlug":"april-grand-prix"`)
	require.Contains(t, rec.Body.String(), `"category":"ATTEMPTED_UNSOLVED"`)
	require.Contains(t, rec.Body.String(), `"problemCode":"D"`)
	require.Contains(t, rec.Body.String(), `"attemptCount":2`)
	require.Contains(t, rec.Body.String(), `"severityRank":3`)
}

type fakeContestMakeupReader struct {
	list contest_makeup.List
	err  error
}

func (f fakeContestMakeupReader) GetMakeupList(ctx context.Context, actor auth.AuthenticatedUser, contestSlug string) (contest_makeup.List, error) {
	if f.err != nil {
		return contest_makeup.List{}, f.err
	}
	return f.list, nil
}

func withAuthenticatedUser(req *http.Request) *http.Request {
	user := auth.AuthenticatedUser{
		User: auth.User{
			ID:       uuid.New(),
			Username: "tester",
			Role:     auth.RoleUser,
		},
	}
	return req.WithContext(context.WithValue(req.Context(), currentUserContextKey{}, user))
}

func withContestSlug(req *http.Request, slug string) *http.Request {
	routeContext := chi.NewRouteContext()
	routeContext.URLParams.Add("slug", slug)
	return req.WithContext(context.WithValue(req.Context(), chi.RouteCtxKey, routeContext))
}
