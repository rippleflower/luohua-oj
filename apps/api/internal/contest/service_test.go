package contest_test

import (
	"context"
	"testing"
	"time"

	"github.com/example/oj3/apps/api/internal/auth"
	"github.com/example/oj3/apps/api/internal/contest"
	"github.com/google/uuid"
	"github.com/stretchr/testify/require"
)

func TestCreateAdminNormalizesAndValidates(t *testing.T) {
	repo := &capturingRepo{}
	service := contest.NewService(repo)
	actor := auth.AuthenticatedUser{User: auth.User{ID: uuid.New(), Role: auth.RoleAdmin}}

	created, err := service.CreateAdmin(context.Background(), actor, contest.CreateAdminInput{
		Slug:        " Spring-Open ",
		Title:       " Spring Open ",
		Description: " intro ",
		Status:      "upcoming",
		StartsAt:    time.Date(2026, 5, 18, 12, 0, 0, 0, time.UTC),
		EndsAt:      time.Date(2026, 5, 18, 14, 0, 0, 0, time.UTC),
		Reason:      " seed ",
	})

	require.NoError(t, err)
	require.Equal(t, "spring-open", repo.created.Slug)
	require.Equal(t, "Spring Open", repo.created.Title)
	require.Equal(t, "intro", repo.created.Description)
	require.Equal(t, "UPCOMING", repo.created.Status)
	require.Equal(t, "seed", repo.created.Reason)
	require.Equal(t, "spring-open", created.Slug)
}

func TestUpdateAdminRejectsInvalidTimeRange(t *testing.T) {
	service := contest.NewService(&capturingRepo{})
	actor := auth.AuthenticatedUser{User: auth.User{ID: uuid.New(), Role: auth.RoleAdmin}}

	_, err := service.UpdateAdmin(context.Background(), actor, contest.UpdateAdminInput{
		ContestID:   uuid.New(),
		Slug:        "spring-open",
		Title:       "Spring Open",
		Description: "intro",
		Status:      "RUNNING",
		StartsAt:    time.Date(2026, 5, 18, 14, 0, 0, 0, time.UTC),
		EndsAt:      time.Date(2026, 5, 18, 12, 0, 0, 0, time.UTC),
	})

	require.ErrorContains(t, err, "endsAt must be after startsAt")
}

func TestReplaceProblemsRejectsDuplicateCodes(t *testing.T) {
	service := contest.NewService(&capturingRepo{})
	actor := auth.AuthenticatedUser{User: auth.User{ID: uuid.New(), Role: auth.RoleAdmin}}

	_, err := service.ReplaceProblemsAdmin(context.Background(), actor, contest.ReplaceProblemsInput{
		ContestID: uuid.New(),
		Problems: []contest.AdminProblemBindingInput{
			{ProblemID: uuid.New(), Code: "A", Position: 1},
			{ProblemID: uuid.New(), Code: "A", Position: 2},
		},
	})

	require.ErrorContains(t, err, "duplicate code")
}

func TestFreezeAdminRejectsMissingContestID(t *testing.T) {
	service := contest.NewService(&capturingRepo{})
	actor := auth.AuthenticatedUser{User: auth.User{ID: uuid.New(), Role: auth.RoleAdmin}}

	_, err := service.FreezeAdmin(context.Background(), actor, contest.FreezeAdminInput{})

	require.ErrorContains(t, err, "contest id is required")
}

type capturingRepo struct {
	created  contest.CreateAdminInput
	updated  contest.UpdateAdminInput
	replaced contest.ReplaceProblemsInput
	frozen   contest.FreezeAdminInput
}

func (c *capturingRepo) ListContests(ctx context.Context) ([]contest.Summary, error) {
	return nil, nil
}

func (c *capturingRepo) GetContestBySlug(ctx context.Context, slug string) (contest.Detail, error) {
	return contest.Detail{}, nil
}

func (c *capturingRepo) ListAdminContests(ctx context.Context) ([]contest.AdminContest, error) {
	return nil, nil
}

func (c *capturingRepo) CreateAdminContest(ctx context.Context, actor auth.AuthenticatedUser, input contest.CreateAdminInput) (contest.AdminContest, error) {
	c.created = input
	return contest.AdminContest{
		ID:        uuid.New(),
		Slug:      input.Slug,
		Title:     input.Title,
		Status:    input.Status,
		StartsAt:  input.StartsAt,
		EndsAt:    input.EndsAt,
		UpdatedAt: time.Now().UTC(),
	}, nil
}

func (c *capturingRepo) UpdateAdminContest(ctx context.Context, actor auth.AuthenticatedUser, input contest.UpdateAdminInput) (contest.AdminContest, error) {
	c.updated = input
	return contest.AdminContest{}, nil
}

func (c *capturingRepo) ReplaceAdminContestProblems(ctx context.Context, actor auth.AuthenticatedUser, input contest.ReplaceProblemsInput) (contest.AdminContest, error) {
	c.replaced = input
	return contest.AdminContest{}, nil
}

func (c *capturingRepo) FreezeAdminContest(ctx context.Context, actor auth.AuthenticatedUser, input contest.FreezeAdminInput) (contest.AdminContest, error) {
	c.frozen = input
	return contest.AdminContest{}, nil
}
