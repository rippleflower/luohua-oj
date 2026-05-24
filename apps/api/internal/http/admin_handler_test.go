package http_test

import (
	"bytes"
	"context"
	"crypto/sha256"
	"encoding/base64"
	"errors"
	"io"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/example/oj3/apps/api/internal/auth"
	"github.com/example/oj3/apps/api/internal/contest"
	apphttp "github.com/example/oj3/apps/api/internal/http"
	"github.com/example/oj3/apps/api/internal/problem"
	"github.com/example/oj3/apps/api/internal/queue"
	"github.com/example/oj3/apps/api/internal/submission"
	"github.com/google/uuid"
	"github.com/stretchr/testify/require"
)

func TestProtectedRoutesRequireAuthentication(t *testing.T) {
	router := testRouter(t, auth.AuthenticatedUser{}, false, &fakeProblemAdmin{}, nil, nil)

	for _, path := range []string{"/me/summary", "/admin/users", "/admin/problems"} {
		req := httptest.NewRequest(http.MethodGet, path, nil)
		rec := httptest.NewRecorder()

		router.ServeHTTP(rec, req)

		require.Equal(t, http.StatusUnauthorized, rec.Code, path)
		require.JSONEq(t, `{"error":"authentication required"}`, rec.Body.String())
	}
}

func TestAdminProblemWritePermissionsAndCSRF(t *testing.T) {
	t.Run("user is forbidden", func(t *testing.T) {
		manager := &fakeProblemAdmin{}
		router := testRouter(t, auth.AuthenticatedUser{
			User: auth.User{
				ID:          uuid.New(),
				Role:        auth.RoleUser,
				DisplayName: "regular",
			},
		}, true, manager, nil, nil)

		req := adminProblemRequest(http.MethodPost, "/admin/problems", `{"slug":"two-sum","title":"Two Sum","difficulty":"EASY","timeLimitMs":1000,"memoryLimitKb":262144}`)
		rec := httptest.NewRecorder()
		router.ServeHTTP(rec, req)

		require.Equal(t, http.StatusForbidden, rec.Code)
		require.JSONEq(t, `{"error":"forbidden"}`, rec.Body.String())
		require.Zero(t, manager.createCalls)
	})

	t.Run("admin without grant is forbidden", func(t *testing.T) {
		manager := &fakeProblemAdmin{}
		router := testRouter(t, auth.AuthenticatedUser{
			User: auth.User{
				ID:          uuid.New(),
				Role:        auth.RoleAdmin,
				DisplayName: "admin",
			},
		}, true, manager, nil, nil)

		req := adminProblemRequest(http.MethodPost, "/admin/problems", `{"slug":"two-sum","title":"Two Sum","difficulty":"EASY","timeLimitMs":1000,"memoryLimitKb":262144}`)
		rec := httptest.NewRecorder()
		router.ServeHTTP(rec, req)

		require.Equal(t, http.StatusForbidden, rec.Code)
		require.JSONEq(t, `{"error":"forbidden"}`, rec.Body.String())
		require.Zero(t, manager.createCalls)
	})

	t.Run("super admin needs csrf", func(t *testing.T) {
		manager := &fakeProblemAdmin{}
		router := testRouter(t, auth.AuthenticatedUser{
			User: auth.User{
				ID:          uuid.New(),
				Role:        auth.RoleSuperAdmin,
				DisplayName: "root",
			},
		}, true, manager, nil, nil)

		req := httptest.NewRequest(http.MethodPost, "/admin/problems", bytes.NewBufferString(`{"slug":"two-sum","title":"Two Sum","difficulty":"EASY","timeLimitMs":1000,"memoryLimitKb":262144}`))
		req.AddCookie(&http.Cookie{Name: "oj_session", Value: testSessionToken()})
		rec := httptest.NewRecorder()
		router.ServeHTTP(rec, req)

		require.Equal(t, http.StatusForbidden, rec.Code)
		require.JSONEq(t, `{"error":"csrf token required"}`, rec.Body.String())
		require.Zero(t, manager.createCalls)
	})
}

func TestAdminProblemWriteRoutesForSuperAdmin(t *testing.T) {
	problemID := uuid.New()
	manager := &fakeProblemAdmin{
		response: problem.AdminProblem{
			ID:               problemID,
			Slug:             "two-sum",
			Title:            "Two Sum",
			Difficulty:       "EASY",
			TimeLimitMs:      1000,
			MemoryLimitKb:    262144,
			Status:           "PUBLISHED",
			CurrentVersionNo: 1,
			IsPublished:      true,
			UpdatedAt:        time.Date(2026, 5, 18, 12, 0, 0, 0, time.UTC),
		},
	}
	router := testRouter(t, auth.AuthenticatedUser{
		User: auth.User{
			ID:          uuid.New(),
			Role:        auth.RoleSuperAdmin,
			DisplayName: "root",
		},
	}, true, manager, nil, nil)

	createReq := adminProblemRequest(http.MethodPost, "/admin/problems", `{"slug":"two-sum","title":"Two Sum","difficulty":"EASY","timeLimitMs":1000,"memoryLimitKb":262144,"reason":"seed"}`)
	createRec := httptest.NewRecorder()
	router.ServeHTTP(createRec, createReq)

	require.Equal(t, http.StatusCreated, createRec.Code)
	require.Equal(t, 1, manager.createCalls)
	require.Equal(t, "two-sum", manager.created.Slug)
	require.Contains(t, createRec.Body.String(), `"timeLimitMs":1000`)

	updateReq := adminProblemRequest(http.MethodPatch, "/admin/problems/"+problemID.String(), `{"slug":"two-sum","title":"Two Sum v2","difficulty":"MEDIUM","timeLimitMs":2000,"memoryLimitKb":524288,"reason":"tune"}`)
	updateRec := httptest.NewRecorder()
	router.ServeHTTP(updateRec, updateReq)

	require.Equal(t, http.StatusOK, updateRec.Code)
	require.Equal(t, 1, manager.updateCalls)
	require.Equal(t, problemID, manager.updated.ProblemID)
	require.Equal(t, "Two Sum v2", manager.updated.Title)
	require.Equal(t, "MEDIUM", manager.updated.Difficulty)

	publishReq := adminProblemRequest(http.MethodPost, "/admin/problems/"+problemID.String()+"/publish", `{"reason":"go live"}`)
	publishRec := httptest.NewRecorder()
	router.ServeHTTP(publishRec, publishReq)

	require.Equal(t, http.StatusOK, publishRec.Code)
	require.Equal(t, 1, manager.publishCalls)
	require.Equal(t, problemID, manager.published.ProblemID)
	require.Equal(t, "go live", manager.published.Reason)
	require.Contains(t, publishRec.Body.String(), `"isPublished":true`)
}

func TestAdminContestWriteRoutesForSuperAdmin(t *testing.T) {
	contestID := uuid.New()
	manager := &fakeContestAdmin{
		response: contest.AdminContest{
			ID:               contestID,
			Slug:             "spring-open",
			Title:            "Spring Open",
			Description:      "intro",
			Status:           "UPCOMING",
			StartsAt:         time.Date(2026, 5, 18, 12, 0, 0, 0, time.UTC),
			EndsAt:           time.Date(2026, 5, 18, 14, 0, 0, 0, time.UTC),
			ProblemCount:     2,
			LatestSnapshotNo: 1,
			UpdatedAt:        time.Date(2026, 5, 18, 12, 0, 0, 0, time.UTC),
			Problems: []contest.AdminProblemBinding{
				{ProblemID: uuid.New(), ProblemSlug: "two-sum", ProblemTitle: "Two Sum", Code: "A", Position: 1},
			},
			Snapshots: []contest.AdminSnapshotSummary{
				{ID: uuid.New(), SnapshotNo: 1, FrozenAt: time.Date(2026, 5, 18, 12, 30, 0, 0, time.UTC), ProblemCount: 2},
			},
		},
	}
	router := testRouter(t, auth.AuthenticatedUser{
		User: auth.User{
			ID:          uuid.New(),
			Role:        auth.RoleSuperAdmin,
			DisplayName: "root",
		},
	}, true, &fakeProblemAdmin{}, manager, &fakeSubmissionAdmin{})

	createReq := adminProblemRequest(http.MethodPost, "/admin/contests", `{"slug":"spring-open","title":"Spring Open","description":"intro","status":"UPCOMING","startsAt":"2026-05-18T12:00:00Z","endsAt":"2026-05-18T14:00:00Z","reason":"seed"}`)
	createRec := httptest.NewRecorder()
	router.ServeHTTP(createRec, createReq)

	require.Equal(t, http.StatusCreated, createRec.Code)
	require.Equal(t, 1, manager.createCalls)
	require.Equal(t, "spring-open", manager.created.Slug)
	require.Contains(t, createRec.Body.String(), `"description":"intro"`)

	updateReq := adminProblemRequest(http.MethodPatch, "/admin/contests/"+contestID.String(), `{"slug":"spring-open","title":"Spring Open v2","description":"intro","status":"RUNNING","startsAt":"2026-05-18T12:00:00Z","endsAt":"2026-05-18T15:00:00Z","reason":"tune"}`)
	updateRec := httptest.NewRecorder()
	router.ServeHTTP(updateRec, updateReq)

	require.Equal(t, http.StatusOK, updateRec.Code)
	require.Equal(t, 1, manager.updateCalls)
	require.Equal(t, contestID, manager.updated.ContestID)
	require.Equal(t, "RUNNING", manager.updated.Status)

	problemsReq := adminProblemRequest(http.MethodPut, "/admin/contests/"+contestID.String()+"/problems", `{"problems":[{"problemId":"`+uuid.New().String()+`","code":"A","position":1}],"reason":"lineup"}`)
	problemsRec := httptest.NewRecorder()
	router.ServeHTTP(problemsRec, problemsReq)

	require.Equal(t, http.StatusOK, problemsRec.Code)
	require.Equal(t, 1, manager.replaceCalls)
	require.Len(t, manager.replaced.Problems, 1)
	require.Equal(t, "A", manager.replaced.Problems[0].Code)

	freezeReq := adminProblemRequest(http.MethodPost, "/admin/contests/"+contestID.String()+"/freeze", `{"reason":"publish"}`)
	freezeRec := httptest.NewRecorder()
	router.ServeHTTP(freezeRec, freezeReq)

	require.Equal(t, http.StatusOK, freezeRec.Code)
	require.Equal(t, 1, manager.freezeCalls)
	require.Equal(t, contestID, manager.frozen.ContestID)
	require.Contains(t, freezeRec.Body.String(), `"latestSnapshotNo":1`)
}

func TestAdminJudgeQueueAndRejudgeRoutes(t *testing.T) {
	submissionID := uuid.New()
	queueSummary := submission.QueueSummary{
		Queue:          "judge",
		Pending:        2,
		Active:         1,
		Retry:          1,
		ProcessedToday: 4,
		RecentTasks: []queue.TaskSummary{
			{ID: "task-1", Type: "judge:submission", SubmissionID: &submissionID, State: "retry"},
		},
		UpdatedAt: time.Date(2026, 5, 18, 12, 0, 0, 0, time.UTC),
	}
	manager := &fakeSubmissionAdmin{
		summary: queueSummary,
		rejudgeResult: submission.RejudgeResult{
			SubmissionID:          submissionID,
			Queue:                 "judge",
			ResultSnapshotVersion: 2,
		},
	}
	router := testRouter(t, auth.AuthenticatedUser{
		User: auth.User{
			ID:          uuid.New(),
			Role:        auth.RoleSuperAdmin,
			DisplayName: "root",
		},
	}, true, &fakeProblemAdmin{}, &fakeContestAdmin{}, manager)

	queueReq := httptest.NewRequest(http.MethodGet, "/admin/judge/queue", nil)
	queueReq.AddCookie(&http.Cookie{Name: "oj_session", Value: testSessionToken()})
	queueRec := httptest.NewRecorder()
	router.ServeHTTP(queueRec, queueReq)

	require.Equal(t, http.StatusOK, queueRec.Code)
	require.Contains(t, queueRec.Body.String(), `"pending":2`)
	require.Contains(t, queueRec.Body.String(), `"submissionId":"`+submissionID.String()+`"`)

	rejudgeReq := adminProblemRequest(http.MethodPost, "/admin/submissions/"+submissionID.String()+"/rejudge", `{"reason":"retry"}`)
	rejudgeRec := httptest.NewRecorder()
	router.ServeHTTP(rejudgeRec, rejudgeReq)

	require.Equal(t, http.StatusOK, rejudgeRec.Code)
	require.Equal(t, 1, manager.rejudgeCalls)
	require.Equal(t, submissionID, manager.rejudged.SubmissionID)
	require.Contains(t, rejudgeRec.Body.String(), `"resultSnapshotVersion":2`)
}

func adminProblemRequest(method string, path string, body string) *http.Request {
	req := httptest.NewRequest(method, path, bytes.NewBufferString(body))
	req.AddCookie(&http.Cookie{Name: "oj_session", Value: testSessionToken()})
	req.AddCookie(&http.Cookie{Name: "oj_csrf", Value: "csrf-token"})
	req.Header.Set("X-CSRF-Token", "csrf-token")
	return req
}

func testRouter(t *testing.T, user auth.AuthenticatedUser, includeSession bool, manager *fakeProblemAdmin, contestManager *fakeContestAdmin, submissionManager *fakeSubmissionAdmin) http.Handler {
	t.Helper()

	repo := &fakeAuthRepo{
		storedUser: auth.StoredUser{
			User: auth.User{
				ID:          user.ID,
				Email:       "admin@example.com",
				Username:    "admin",
				DisplayName: user.DisplayName,
				Role:        user.Role,
				Permissions: user.Permissions,
				Status:      "ACTIVE",
			},
			PasswordHash: "",
		},
	}
	if includeSession {
		authenticated := user
		authenticated.Email = "admin@example.com"
		authenticated.Username = "admin"
		authenticated.Status = "ACTIVE"
		authenticated.SessionID = uuid.New()
		authenticated.SessionTokenHash = testSessionHash()
		repo.authenticated = authenticated
	}

	authService := auth.NewService(repo, time.Hour, slog.New(slog.NewJSONHandler(io.Discard, nil)))
	if contestManager == nil {
		contestManager = &fakeContestAdmin{}
	}
	if submissionManager == nil {
		submissionManager = &fakeSubmissionAdmin{}
	}
	return apphttp.NewRouter(apphttp.RouterOptions{
		AuthService:       authService,
		ProblemAdmin:      manager,
		ContestAdmin:      contestManager,
		SubmissionAdmin:   submissionManager,
		SessionCookieName: "oj_session",
	})
}

func testSessionToken() string {
	return base64.RawURLEncoding.EncodeToString([]byte("0123456789abcdef0123456789abcdef"))
}

func testSessionHash() string {
	sum := sha256.Sum256([]byte(testSessionToken()))
	return base64.RawURLEncoding.EncodeToString(sum[:])
}

type fakeProblemAdmin struct {
	response     problem.AdminProblem
	created      problem.CreateAdminInput
	updated      problem.UpdateAdminInput
	published    problem.PublishAdminInput
	createCalls  int
	updateCalls  int
	publishCalls int
}

type fakeContestAdmin struct {
	response     contest.AdminContest
	created      contest.CreateAdminInput
	updated      contest.UpdateAdminInput
	replaced     contest.ReplaceProblemsInput
	frozen       contest.FreezeAdminInput
	createCalls  int
	updateCalls  int
	replaceCalls int
	freezeCalls  int
}

func (f *fakeContestAdmin) ListAdmin(ctx context.Context) ([]contest.AdminContest, error) {
	if f.response.ID == uuid.Nil {
		return nil, nil
	}
	return []contest.AdminContest{f.response}, nil
}

func (f *fakeContestAdmin) CreateAdmin(ctx context.Context, actor auth.AuthenticatedUser, input contest.CreateAdminInput) (contest.AdminContest, error) {
	f.createCalls++
	f.created = input
	return f.response, nil
}

func (f *fakeContestAdmin) UpdateAdmin(ctx context.Context, actor auth.AuthenticatedUser, input contest.UpdateAdminInput) (contest.AdminContest, error) {
	f.updateCalls++
	f.updated = input
	return f.response, nil
}

func (f *fakeContestAdmin) ReplaceProblemsAdmin(ctx context.Context, actor auth.AuthenticatedUser, input contest.ReplaceProblemsInput) (contest.AdminContest, error) {
	f.replaceCalls++
	f.replaced = input
	return f.response, nil
}

func (f *fakeContestAdmin) FreezeAdmin(ctx context.Context, actor auth.AuthenticatedUser, input contest.FreezeAdminInput) (contest.AdminContest, error) {
	f.freezeCalls++
	f.frozen = input
	return f.response, nil
}

type fakeSubmissionAdmin struct {
	summary       submission.QueueSummary
	rejudgeResult submission.RejudgeResult
	rejudged      submission.RejudgeAdminInput
	rejudgeCalls  int
}

func (f *fakeSubmissionAdmin) RejudgeAdmin(ctx context.Context, actor auth.AuthenticatedUser, input submission.RejudgeAdminInput) (submission.RejudgeResult, error) {
	f.rejudgeCalls++
	f.rejudged = input
	return f.rejudgeResult, nil
}

func (f *fakeSubmissionAdmin) QueueSummary(ctx context.Context) (submission.QueueSummary, error) {
	return f.summary, nil
}

func (f *fakeProblemAdmin) CreateAdmin(ctx context.Context, actor auth.AuthenticatedUser, input problem.CreateAdminInput) (problem.AdminProblem, error) {
	f.createCalls++
	f.created = input
	return f.response, nil
}

func (f *fakeProblemAdmin) UpdateAdmin(ctx context.Context, actor auth.AuthenticatedUser, input problem.UpdateAdminInput) (problem.AdminProblem, error) {
	f.updateCalls++
	f.updated = input
	return f.response, nil
}

func (f *fakeProblemAdmin) PublishAdmin(ctx context.Context, actor auth.AuthenticatedUser, input problem.PublishAdminInput) (problem.AdminProblem, error) {
	f.publishCalls++
	f.published = input
	return f.response, nil
}

type fakeAuthRepo struct {
	authenticated auth.AuthenticatedUser
	storedUser    auth.StoredUser
}

func (f *fakeAuthRepo) CreateUser(ctx context.Context, email string, username string, passwordHash string, displayName string) (auth.User, error) {
	return auth.User{}, nil
}

func (f *fakeAuthRepo) FindUserByIdentifier(ctx context.Context, identifier string) (auth.StoredUser, error) {
	return f.storedUser, nil
}

func (f *fakeAuthRepo) GetStoredUserByID(ctx context.Context, userID uuid.UUID) (auth.StoredUser, error) {
	return f.storedUser, nil
}

func (f *fakeAuthRepo) CreateSession(ctx context.Context, userID uuid.UUID, tokenHash string, ip string, userAgent string, expiresAt time.Time) (uuid.UUID, error) {
	return uuid.New(), nil
}

func (f *fakeAuthRepo) GetAuthenticatedUserBySessionHash(ctx context.Context, tokenHash string) (auth.AuthenticatedUser, error) {
	if tokenHash != testSessionHash() || f.authenticated.ID == uuid.Nil {
		return auth.AuthenticatedUser{}, errors.New("missing session")
	}
	return f.authenticated, nil
}

func (f *fakeAuthRepo) TouchSession(ctx context.Context, sessionID uuid.UUID) error {
	return nil
}

func (f *fakeAuthRepo) DeleteSessionByHash(ctx context.Context, tokenHash string) error {
	return nil
}

func (f *fakeAuthRepo) ListSessions(ctx context.Context, userID uuid.UUID, currentHash string) ([]auth.Session, error) {
	return nil, nil
}

func (f *fakeAuthRepo) RevokeSession(ctx context.Context, userID uuid.UUID, sessionID uuid.UUID) error {
	return nil
}

func (f *fakeAuthRepo) UpdatePassword(ctx context.Context, userID uuid.UUID, passwordHash string) error {
	return nil
}

func (f *fakeAuthRepo) RevokeOtherSessions(ctx context.Context, userID uuid.UUID, currentHash string) error {
	return nil
}

func (f *fakeAuthRepo) GetMeSummary(ctx context.Context, userID uuid.UUID) (auth.MeSummary, error) {
	return auth.MeSummary{}, nil
}

func (f *fakeAuthRepo) GetMeSettings(ctx context.Context, userID uuid.UUID, currentHash string) (auth.MeSettings, error) {
	return auth.MeSettings{}, nil
}

func (f *fakeAuthRepo) UpdateProfile(ctx context.Context, input auth.UpdateProfileInput) (auth.MeSettings, error) {
	return auth.MeSettings{}, nil
}

func (f *fakeAuthRepo) UpdatePreferences(ctx context.Context, input auth.UpdatePreferencesInput) (auth.MeSettings, error) {
	return auth.MeSettings{}, nil
}

func (f *fakeAuthRepo) GetAdminDashboard(ctx context.Context, actor auth.User) (auth.Dashboard, error) {
	return auth.Dashboard{}, nil
}

func (f *fakeAuthRepo) ListAdminUsers(ctx context.Context) ([]auth.AdminUserSummary, error) {
	return nil, nil
}

func (f *fakeAuthRepo) GetAdminUser(ctx context.Context, targetID uuid.UUID) (auth.AdminUserDetail, error) {
	return auth.AdminUserDetail{}, nil
}

func (f *fakeAuthRepo) GetAdminPermissions(ctx context.Context, targetID uuid.UUID) ([]auth.PermissionKey, error) {
	return nil, nil
}

func (f *fakeAuthRepo) UpdateAdminUser(ctx context.Context, input auth.UpdateAdminUserInput) (auth.AdminUserDetail, error) {
	return auth.AdminUserDetail{}, nil
}

func (f *fakeAuthRepo) SetUserRole(ctx context.Context, actor auth.AuthenticatedUser, targetID uuid.UUID, role auth.Role, reason string, ip string) error {
	return nil
}

func (f *fakeAuthRepo) ReplacePermissions(ctx context.Context, actor auth.AuthenticatedUser, targetID uuid.UUID, permissions []auth.PermissionKey, reason string, ip string) error {
	return nil
}

func (f *fakeAuthRepo) ListAdminProblems(ctx context.Context) ([]auth.AdminProblemSummary, error) {
	return nil, nil
}

func (f *fakeAuthRepo) ListAdminContests(ctx context.Context) ([]auth.AdminContestSummary, error) {
	return nil, nil
}

func (f *fakeAuthRepo) ListAdminSubmissions(ctx context.Context) ([]auth.AdminSubmissionSummary, error) {
	return nil, nil
}

func (f *fakeAuthRepo) ListAuditEvents(ctx context.Context) ([]auth.AuditEvent, error) {
	return nil, nil
}

func (f *fakeAuthRepo) GetSystemSettings(ctx context.Context, sourceRoot string, redisAddr string) (auth.SystemSettings, error) {
	return auth.SystemSettings{}, nil
}

func (f *fakeAuthRepo) UpdateSystemSettings(ctx context.Context, actor auth.AuthenticatedUser, settings auth.SystemSettings, ip string) (auth.SystemSettings, error) {
	return auth.SystemSettings{}, nil
}

func (f *fakeAuthRepo) ListAnnouncements(ctx context.Context) ([]auth.Announcement, error) {
	return nil, nil
}

func (f *fakeAuthRepo) CreateAnnouncement(ctx context.Context, actor auth.AuthenticatedUser, title string, content string, status string, audience string, ip string) (auth.Announcement, error) {
	return auth.Announcement{}, nil
}

func (f *fakeAuthRepo) UpdateAnnouncement(ctx context.Context, actor auth.AuthenticatedUser, announcementID uuid.UUID, title string, content string, status string, audience string, ip string) (auth.Announcement, error) {
	return auth.Announcement{}, nil
}
