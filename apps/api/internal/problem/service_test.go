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
	service := problem.NewService(repo, problem.NewRouteCodec("test-salt"))
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
	service := problem.NewService(&capturingRepo{}, problem.NewRouteCodec("test-salt"))
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
	service := problem.NewService(&capturingRepo{}, problem.NewRouteCodec("test-salt"))
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
	service := problem.NewService(&capturingRepo{}, problem.NewRouteCodec("test-salt"))
	actor := auth.AuthenticatedUser{User: auth.User{ID: uuid.New(), Role: auth.RoleAdmin}}

	_, err := service.PublishAdmin(context.Background(), actor, problem.PublishAdminInput{})

	require.ErrorContains(t, err, "problem id is required")
}

func TestUpdateAdminContentNormalizesSectionsAndTags(t *testing.T) {
	repo := &capturingRepo{}
	service := problem.NewService(repo, problem.NewRouteCodec("test-salt"))
	actor := auth.AuthenticatedUser{User: auth.User{ID: uuid.New(), Role: auth.RoleAdmin}}
	problemID := uuid.New()

	updated, err := service.UpdateAdminContent(context.Background(), actor, problem.UpdateAdminContentInput{
		ProblemID: problemID,
		StatementJSON: []problem.StatementSection{
			{Kind: "MARKDOWN", Section: "constraints", Content: " limit "},
			{Kind: "", Section: "statement", Content: " body "},
			{Kind: "markdown", Section: "output", Content: " out "},
			{Kind: "markdown", Section: "input", Content: " in "},
		},
		Samples: []problem.Sample{
			{Input: "1 2\r\n", Output: "3\r\n", Weight: 1},
		},
		Tags:   []string{" Graph ", "graph", " shortest-path "},
		Reason: " tune ",
	})

	require.NoError(t, err)
	require.Equal(t, problemID, repo.contentUpdated.ProblemID)
	require.Equal(t, []string{"graph", "shortest-path"}, repo.contentUpdated.Tags)
	require.Equal(t, []string{"statement", "input", "output", "constraints"}, []string{
		repo.contentUpdated.StatementJSON[0].Section,
		repo.contentUpdated.StatementJSON[1].Section,
		repo.contentUpdated.StatementJSON[2].Section,
		repo.contentUpdated.StatementJSON[3].Section,
	})
	require.Equal(t, "markdown", repo.contentUpdated.StatementJSON[0].Kind)
	require.Equal(t, "1 2\n", repo.contentUpdated.Samples[0].Input)
	require.Equal(t, "graph", updated.Tags[0])
}

type capturingRepo struct {
	created        problem.CreateAdminInput
	updated        problem.UpdateAdminInput
	contentUpdated problem.UpdateAdminContentInput
	published      problem.PublishAdminInput
}

func (c *capturingRepo) ListProblems(ctx context.Context) ([]problem.Summary, error) {
	return nil, nil
}

func (c *capturingRepo) GetProblemBySlug(ctx context.Context, slug string) (problem.Detail, error) {
	return problem.Detail{}, nil
}

func (c *capturingRepo) GetProblemByNumber(ctx context.Context, problemNo int64) (problem.Detail, error) {
	return problem.Detail{ProblemNo: problemNo}, nil
}

func (c *capturingRepo) GetAdminProblemDetail(ctx context.Context, problemID uuid.UUID) (problem.AdminProblemDetail, error) {
	return problem.AdminProblemDetail{
		AdminProblem: problem.AdminProblem{
			ID:        problemID,
			ProblemNo: 1,
		},
	}, nil
}

func (c *capturingRepo) CreateAdminProblem(ctx context.Context, actor auth.AuthenticatedUser, input problem.CreateAdminInput) (problem.AdminProblem, error) {
	c.created = input
	return problem.AdminProblem{
		ID:               uuid.New(),
		ProblemNo:        1,
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
	return problem.AdminProblem{ProblemNo: 1}, nil
}

func (c *capturingRepo) PublishAdminProblem(ctx context.Context, actor auth.AuthenticatedUser, input problem.PublishAdminInput) (problem.AdminProblem, error) {
	c.published = input
	return problem.AdminProblem{ProblemNo: 1}, nil
}

func (c *capturingRepo) UpdateAdminProblemContent(ctx context.Context, actor auth.AuthenticatedUser, input problem.UpdateAdminContentInput) (problem.AdminProblemDetail, error) {
	c.contentUpdated = input
	return problem.AdminProblemDetail{
		AdminProblem: problem.AdminProblem{
			ID:        input.ProblemID,
			ProblemNo: 1,
		},
		StatementJSON: input.StatementJSON,
		Samples:       input.Samples,
		Tags:          input.Tags,
	}, nil
}
