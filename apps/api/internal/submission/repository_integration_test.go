package submission_test

import (
	"context"
	"os"
	"testing"
	"time"

	db "github.com/example/oj3/apps/api/internal/db/generated"
	"github.com/example/oj3/apps/api/internal/submission"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/stretchr/testify/require"
)

func TestSQLRepositoryCreateSubmission(t *testing.T) {
	databaseURL := os.Getenv("DATABASE_URL")
	if databaseURL == "" {
		databaseURL = "postgres://oj:oj@localhost:5432/oj?sslmode=disable"
	}

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	conn, err := pgx.Connect(ctx, databaseURL)
	require.NoError(t, err)
	defer conn.Close(context.Background())

	queries := db.New(conn)
	suffix := time.Now().Format("20060102150405000000000")
	user, err := queries.CreateUser(ctx, db.CreateUserParams{
		Email:        "repo-test-" + suffix + "@example.com",
		Username:     "repo_test_user_" + suffix,
		PasswordHash: "hash",
		Role:         db.UserRoleUSER,
	})
	require.NoError(t, err)

	problem, err := queries.CreateProblem(ctx, db.CreateProblemParams{
		Slug:          "repo-test-problem-" + suffix,
		Title:         "Repository Test Problem",
		Difficulty:    db.ProblemDifficultyEASY,
		StatementMd:   "statement",
		InputMd:       "input",
		OutputMd:      "output",
		ConstraintsMd: "constraints",
		TimeLimitMs:   1000,
		MemoryLimitKb: 262144,
		IsPublished:   true,
	})
	require.NoError(t, err)

	repo := submission.NewSQLRepository(queries)
	submissionID := uuid.New()
	created, err := repo.CreateSubmission(ctx, submission.CreateParams{
		ID:           submissionID,
		UserID:       uuid.UUID(user.ID.Bytes),
		ProblemID:    uuid.UUID(problem.ID.Bytes),
		Language:     "CPP17",
		SourceObject: "tmp/submissions/repo-test.cpp",
	})

	require.NoError(t, err)
	require.Equal(t, submissionID, created.ID)
	require.Equal(t, submission.StatusPending, created.Status)
	require.Equal(t, "CPP17", created.Language)
}
