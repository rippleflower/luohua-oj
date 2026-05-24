package problem_test

import (
	"context"
	"testing"
	"time"

	"github.com/example/oj3/apps/api/internal/auth"
	"github.com/example/oj3/apps/api/internal/problem"
	"github.com/google/uuid"
	"github.com/stretchr/testify/require"
)

func TestCreateAdminNormalizesAndValidates(t *testing.T) {
	repo := &capturingRepo{}
	service := problem.NewService(repo)
	actor := auth.AuthenticatedUser{User: auth.User{ID: uuid.New(), Role: auth.RoleAdmin}}

	created, err := service.CreateAdmin(context.Background(), actor, problem.CreateAdminInput{
		Slug:          "  Two-Sum  ",
		Title:         "  Two Sum  ",
		Difficulty:    "easy",
		TimeLimitMs:   1000,
		MemoryLimitKb: 262144,
		Reason:        "  initial draft  ",
	})

	require.NoError(t, err)
	require.Equal(t, "two-sum", repo.created.Slug)
	require.Equal(t, "Two Sum", repo.created.Title)
	require.Equal(t, "EASY", repo.created.Difficulty)
	require.Equal(t, "initial draft", repo.created.Reason)
	require.Equal(t, "two-sum", created.Slug)
}

func TestCreateAdminRejectsInvalidDifficulty(t *testing.T) {
	service := problem.NewService(&capturingRepo{})
	actor := auth.AuthenticatedUser{User: auth.User{ID: uuid.New(), Role: auth.RoleAdmin}}

	_, err := service.CreateAdmin(context.Background(), actor, problem.CreateAdminInput{
		Slug:          "two-sum",
		Title:         "Two Sum",
		Difficulty:    "IMPOSSIBLE",
		TimeLimitMs:   1000,
		MemoryLimitKb: 262144,
	})

	require.ErrorContains(t, err, "invalid difficulty")
}

func TestUpdateAdminRejectsMissingProblemID(t *testing.T) {
	service := problem.NewService(&capturingRepo{})
	actor := auth.AuthenticatedUser{User: auth.User{ID: uuid.New(), Role: auth.RoleAdmin}}

	_, err := service.UpdateAdmin(context.Background(), actor, problem.UpdateAdminInput{
		Slug:          "two-sum",
		Title:         "Two Sum",
		Difficulty:    "EASY",
		TimeLimitMs:   1000,
		MemoryLimitKb: 262144,
	})

	require.ErrorContains(t, err, "problem id is required")
}

func TestPublishAdminRejectsMissingProblemID(t *testing.T) {
	service := problem.NewService(&capturingRepo{})
	actor := auth.AuthenticatedUser{User: auth.User{ID: uuid.New(), Role: auth.RoleAdmin}}

	_, err := service.PublishAdmin(context.Background(), actor, problem.PublishAdminInput{})

	require.ErrorContains(t, err, "problem id is required")
}

type capturingRepo struct {
	created   problem.CreateAdminInput
	updated   problem.UpdateAdminInput
	published problem.PublishAdminInput
}

func (c *capturingRepo) ListProblems(ctx context.Context) ([]problem.Summary, error) {
	return nil, nil
}

func (c *capturingRepo) GetProblemBySlug(ctx context.Context, slug string) (problem.Detail, error) {
	return problem.Detail{}, nil
}

func (c *capturingRepo) CreateAdminProblem(ctx context.Context, actor auth.AuthenticatedUser, input problem.CreateAdminInput) (problem.AdminProblem, error) {
	c.created = input
	return problem.AdminProblem{
		ID:               uuid.New(),
		Slug:             input.Slug,
		Title:            input.Title,
		Difficulty:       input.Difficulty,
		TimeLimitMs:      input.TimeLimitMs,
		MemoryLimitKb:    input.MemoryLimitKb,
		Status:           "DRAFT",
		CurrentVersionNo: 1,
		UpdatedAt:        time.Now().UTC(),
	}, nil
}

func (c *capturingRepo) UpdateAdminProblem(ctx context.Context, actor auth.AuthenticatedUser, input problem.UpdateAdminInput) (problem.AdminProblem, error) {
	c.updated = input
	return problem.AdminProblem{}, nil
}

func (c *capturingRepo) PublishAdminProblem(ctx context.Context, actor auth.AuthenticatedUser, input problem.PublishAdminInput) (problem.AdminProblem, error) {
	c.published = input
	return problem.AdminProblem{}, nil
}
