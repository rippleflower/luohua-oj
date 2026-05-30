package submission_test

import (
	"context"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	db "github.com/example/oj3/apps/api/internal/db/generated"
	"github.com/example/oj3/apps/api/internal/submission"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgtype"
	"github.com/stretchr/testify/require"
)

func TestSQLRepositoryCreateSubmission(t *testing.T) {
	ctx, conn := testDB(t)
	queries := db.New(conn)
	fixture := newSubmissionFixture(t, ctx, queries)

	repo := submission.NewSQLRepository(conn)
	submissionID := uuid.New()
	created, err := repo.CreateSubmission(ctx, submission.CreateParams{
		ID:              submissionID,
		UserID:          fixture.userID,
		ProblemID:       fixture.problemID,
		Language:        "CPP17",
		SourceObjectKey: "submissions/2026/05/repo-test/source.cpp.zst",
	})

	require.NoError(t, err)
	require.Equal(t, submissionID, created.ID)
	require.Equal(t, submission.StatusPending, created.Status)
	require.Equal(t, "CPP17", created.Language)
	require.False(t, created.CreatedAt.IsZero())
}

func TestSQLRepositoryUpdateSubmissionStatusRejectsNegativeMetrics(t *testing.T) {
	ctx, conn := testDB(t)
	queries := db.New(conn)
	fixture := newSubmissionFixture(t, ctx, queries)

	_, err := queries.UpdateSubmissionStatus(ctx, db.UpdateSubmissionStatusParams{
		ID:            pgUUID(fixture.submissionID),
		Status:        db.SubmissionStatusACCEPTED,
		Score:         100,
		CompileOutput: pgtype.Text{},
		MaxTimeMs:     pgtype.Int4{Int32: -1, Valid: true},
		MaxMemoryKb:   pgtype.Int4{Int32: 64, Valid: true},
	})

	require.Error(t, err)
	require.Contains(t, err.Error(), "submissions_max_time_ms_nonnegative")
}

func TestSubmissionResultsRejectDuplicateTestCaseRows(t *testing.T) {
	ctx, conn := testDB(t)
	queries := db.New(conn)
	fixture := newSubmissionFixture(t, ctx, queries)
	testCaseID := uuid.New()

	_, err := conn.Exec(ctx,
		`INSERT INTO test_cases (id, problem_id, input_object, output_object, is_sample, weight)
		 VALUES ($1, $2, $3, $4, $5, $6)`,
		pgUUID(testCaseID),
		pgUUID(fixture.problemID),
		"tmp/testcases/in.txt",
		"tmp/testcases/out.txt",
		true,
		1,
	)
	require.NoError(t, err)

	_, err = conn.Exec(ctx,
		`INSERT INTO submission_results (submission_id, test_case_id, status, time_ms, memory_kb)
		 VALUES ($1, $2, $3, $4, $5)`,
		pgUUID(fixture.submissionID),
		pgUUID(testCaseID),
		db.SubmissionStatusACCEPTED,
		12,
		256,
	)
	require.NoError(t, err)

	_, err = conn.Exec(ctx,
		`INSERT INTO submission_results (submission_id, test_case_id, status, time_ms, memory_kb)
		 VALUES ($1, $2, $3, $4, $5)`,
		pgUUID(fixture.submissionID),
		pgUUID(testCaseID),
		db.SubmissionStatusACCEPTED,
		15,
		512,
	)

	require.Error(t, err)
	require.Contains(t, err.Error(), "submission_results_submission_test_case_unique")
}

func TestSubmissionResultsRejectNegativeMetrics(t *testing.T) {
	ctx, conn := testDB(t)
	queries := db.New(conn)
	fixture := newSubmissionFixture(t, ctx, queries)
	testCaseID := uuid.New()

	_, err := conn.Exec(ctx,
		`INSERT INTO test_cases (id, problem_id, input_object, output_object, is_sample, weight)
		 VALUES ($1, $2, $3, $4, $5, $6)`,
		pgUUID(testCaseID),
		pgUUID(fixture.problemID),
		"tmp/testcases/neg-in.txt",
		"tmp/testcases/neg-out.txt",
		false,
		1,
	)
	require.NoError(t, err)

	_, err = conn.Exec(ctx,
		`INSERT INTO submission_results (submission_id, test_case_id, status, time_ms, memory_kb)
		 VALUES ($1, $2, $3, $4, $5)`,
		pgUUID(fixture.submissionID),
		pgUUID(testCaseID),
		db.SubmissionStatusRUNTIMEERROR,
		-1,
		256,
	)

	require.Error(t, err)
	require.Contains(t, err.Error(), "submission_results_time_ms_nonnegative")
}

func TestSubmissionResultsAllowNullMetrics(t *testing.T) {
	ctx, conn := testDB(t)
	queries := db.New(conn)
	fixture := newSubmissionFixture(t, ctx, queries)
	testCaseID := uuid.New()

	_, err := conn.Exec(ctx,
		`INSERT INTO test_cases (id, problem_id, input_object, output_object, is_sample, weight)
		 VALUES ($1, $2, $3, $4, $5, $6)`,
		pgUUID(testCaseID),
		pgUUID(fixture.problemID),
		"tmp/testcases/null-in.txt",
		"tmp/testcases/null-out.txt",
		false,
		1,
	)
	require.NoError(t, err)

	_, err = conn.Exec(ctx,
		`INSERT INTO submission_results (submission_id, test_case_id, status, time_ms, memory_kb)
		 VALUES ($1, $2, $3, $4, $5)`,
		pgUUID(fixture.submissionID),
		pgUUID(testCaseID),
		db.SubmissionStatusPENDING,
		nil,
		nil,
	)

	require.NoError(t, err)
}

func TestSQLRepositoryListSubmissionsByUsername(t *testing.T) {
	ctx, conn := testDB(t)
	queries := db.New(conn)
	fixture := newSubmissionFixture(t, ctx, queries)

	olderID := uuid.New()
	olderCreatedAt := time.Now().UTC().Add(-2 * time.Hour)
	_, err := conn.Exec(ctx,
		`INSERT INTO submissions (id, user_id, problem_id, language, source_object, status, created_at)
		 VALUES ($1, $2, $3, $4, $5, $6, $7)`,
		pgUUID(olderID),
		pgUUID(fixture.userID),
		pgUUID(fixture.problemID),
		db.LanguageCPP17,
		"tmp/submissions/older.cpp",
		db.SubmissionStatusPENDING,
		olderCreatedAt,
	)
	require.NoError(t, err)

	newerID := uuid.New()
	newerCreatedAt := time.Now().UTC().Add(-1 * time.Hour)
	_, err = conn.Exec(ctx,
		`INSERT INTO submissions (id, user_id, problem_id, language, source_object, status, created_at)
		 VALUES ($1, $2, $3, $4, $5, $6, $7)`,
		pgUUID(newerID),
		pgUUID(fixture.userID),
		pgUUID(fixture.problemID),
		db.LanguageCPP20,
		"tmp/submissions/newer.cpp",
		db.SubmissionStatusACCEPTED,
		newerCreatedAt,
	)
	require.NoError(t, err)

	repo := submission.NewSQLRepository(conn)
	require.NoError(t, repo.RefreshSubmissionViews(ctx, fixture.submissionID))
	require.NoError(t, repo.RefreshSubmissionViews(ctx, newerID))
	require.NoError(t, repo.RefreshSubmissionViews(ctx, olderID))

	firstPage, err := repo.ListSubmissionsByUsername(ctx, submission.ListByUsernameParams{
		Username: fixture.username,
		Page:     1,
		PageSize: 1,
	})
	require.NoError(t, err)
	require.Equal(t, 3, firstPage.Total)
	require.Equal(t, 1, firstPage.Page)
	require.Equal(t, 1, firstPage.PageSize)
	require.Len(t, firstPage.Items, 1)
	require.Equal(t, fixture.submissionID, firstPage.Items[0].ID)
	require.NotNil(t, firstPage.Items[0].Problem)
	require.Equal(t, fixture.problemID, firstPage.Items[0].Problem.ID)
	require.Equal(t, fixture.problemSlug, firstPage.Items[0].Problem.Slug)
	require.Equal(t, fixture.problemTitle, firstPage.Items[0].Problem.Title)

	secondPage, err := repo.ListSubmissionsByUsername(ctx, submission.ListByUsernameParams{
		Username: fixture.username,
		Page:     2,
		PageSize: 1,
	})
	require.NoError(t, err)
	require.Equal(t, 3, secondPage.Total)
	require.Len(t, secondPage.Items, 1)
	require.Equal(t, newerID, secondPage.Items[0].ID)

	thirdPage, err := repo.ListSubmissionsByUsername(ctx, submission.ListByUsernameParams{
		Username: fixture.username,
		Page:     3,
		PageSize: 1,
	})
	require.NoError(t, err)
	require.Equal(t, 3, thirdPage.Total)
	require.Len(t, thirdPage.Items, 1)
	require.Equal(t, olderID, thirdPage.Items[0].ID)

	emptyPage, err := repo.ListSubmissionsByUsername(ctx, submission.ListByUsernameParams{
		Username: fixture.username,
		Page:     4,
		PageSize: 1,
	})
	require.NoError(t, err)
	require.Equal(t, 0, emptyPage.Total)
	require.Empty(t, emptyPage.Items)
}

type submissionFixture struct {
	username     string
	userID       uuid.UUID
	problemID    uuid.UUID
	problemSlug  string
	problemTitle string
	submissionID uuid.UUID
}

func testDB(t *testing.T) (context.Context, *pgx.Conn) {
	t.Helper()

	databaseURL := os.Getenv("DATABASE_URL")
	if databaseURL == "" {
		databaseURL = "postgres://oj:oj@localhost:5432/oj?sslmode=disable"
	}

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	t.Cleanup(cancel)

	conn, err := pgx.Connect(ctx, databaseURL)
	if err != nil {
		t.Skipf("skipping integration test: database unavailable at %q: %v", databaseURL, err)
	}
	t.Cleanup(func() {
		_ = conn.Close(context.Background())
	})

	applyConstraintMigration(t, ctx, conn)
	applyStorageReadModelMigration(t, ctx, conn)
	return ctx, conn
}

func applyConstraintMigration(t *testing.T, ctx context.Context, conn *pgx.Conn) {
	t.Helper()

	migrationPath := filepath.Join("..", "..", "..", "..", "packages", "database", "migrations", "000002_submission_constraints.sql")
	content, err := os.ReadFile(migrationPath)
	require.NoError(t, err)

	parts := strings.Split(string(content), "-- +goose Down")
	require.Len(t, parts, 2)

	upSQL := strings.TrimSpace(strings.TrimPrefix(parts[0], "-- +goose Up"))
	_, err = conn.Exec(ctx, upSQL)
	if err != nil && !strings.Contains(err.Error(), "already exists") {
		require.NoError(t, err)
	}
}

func applyStorageReadModelMigration(t *testing.T, ctx context.Context, conn *pgx.Conn) {
	t.Helper()

	migrationPath := filepath.Join("..", "..", "..", "..", "packages", "database", "migrations", "000003_storage_read_models.sql")
	content, err := os.ReadFile(migrationPath)
	require.NoError(t, err)

	parts := strings.Split(string(content), "-- +goose Down")
	require.Len(t, parts, 2)

	upSQL := strings.TrimSpace(strings.TrimPrefix(parts[0], "-- +goose Up"))
	_, err = conn.Exec(ctx, upSQL)
	if err != nil && !strings.Contains(err.Error(), "already exists") && !strings.Contains(err.Error(), "already a member") {
		require.NoError(t, err)
	}
}

func newSubmissionFixture(t *testing.T, ctx context.Context, queries *db.Queries) submissionFixture {
	t.Helper()

	suffix := strings.ReplaceAll(uuid.NewString(), "-", "")
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

	submissionID := uuid.New()
	_, err = queries.CreateSubmission(ctx, db.CreateSubmissionParams{
		ID:           pgUUID(submissionID),
		UserID:       user.ID,
		ProblemID:    problem.ID,
		Language:     db.LanguageCPP17,
		SourceObject: "submissions/2026/05/fixture/source.cpp.zst",
	})
	require.NoError(t, err)

	return submissionFixture{
		username:     user.Username,
		userID:       uuid.UUID(user.ID.Bytes),
		problemID:    uuid.UUID(problem.ID.Bytes),
		problemSlug:  problem.Slug,
		problemTitle: problem.Title,
		submissionID: submissionID,
	}
}

func pgUUID(id uuid.UUID) pgtype.UUID {
	return pgtype.UUID{Bytes: id, Valid: true}
}
