package contest_makeup_test

import (
	"context"
	"fmt"
	"math"
	"os"
	"path/filepath"
	"regexp"
	"sort"
	"strings"
	"testing"
	"time"

	"github.com/example/oj3/apps/api/internal/contest_makeup"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/stretchr/testify/require"
)

func TestSQLRepositoryListUserProblemAttempts(t *testing.T) {
	ctx, pool := testDB(t)
	repo := contest_makeup.NewSQLRepository(pool)
	fixture := seedContestFixture(t, ctx, pool)

	problems, err := repo.ListContestProblems(ctx, fixture.contestID)
	require.NoError(t, err)
	require.Len(t, problems, 3)
	require.Equal(t, "A", problems[0].ProblemCode)
	require.Equal(t, "B", problems[1].ProblemCode)
	require.Equal(t, "C", problems[2].ProblemCode)

	attempts, err := repo.ListUserProblemAttempts(ctx, fixture.userID, fixture.contestID)
	require.NoError(t, err)
	require.Len(t, attempts, 2)

	problemAAttempt := attempts[fixture.problemA]
	require.Equal(t, "TIME_LIMIT_EXCEEDED", problemAAttempt.LatestStatus)
	require.Equal(t, 2, problemAAttempt.AttemptCount)
	require.False(t, problemAAttempt.HasAccepted)

	problemBAttempt := attempts[fixture.problemB]
	require.Equal(t, "ACCEPTED", problemBAttempt.LatestStatus)
	require.Equal(t, 1, problemBAttempt.AttemptCount)
	require.True(t, problemBAttempt.HasAccepted)
}

func TestSQLRepositoryAttemptQueryExplainUsesIndexPath(t *testing.T) {
	ctx, pool := testDB(t)
	fixture := seedContestFixture(t, ctx, pool)

	conn, err := pool.Acquire(ctx)
	require.NoError(t, err)
	defer conn.Release()

	_, err = conn.Exec(ctx, `SET enable_seqscan = off`)
	require.NoError(t, err)

	rows, err := conn.Query(
		ctx,
		`EXPLAIN (FORMAT TEXT)
		 WITH filtered AS (
		   SELECT s.problem_id, s.status, s.created_at, s.id
		   FROM submissions s
		   WHERE s.user_id = $1
		     AND s.contest_id = $2
		 ),
		 aggregated AS (
		   SELECT f.problem_id, BOOL_OR(f.status = 'ACCEPTED') AS has_accepted, COUNT(*)::int AS attempt_count
		   FROM filtered f
		   GROUP BY f.problem_id
		 ),
		 latest AS (
		   SELECT DISTINCT ON (f.problem_id) f.problem_id, f.status::text AS latest_status
		   FROM filtered f
		   ORDER BY f.problem_id, f.created_at DESC, f.id DESC
		 )
		 SELECT l.problem_id, l.latest_status, a.has_accepted, a.attempt_count
		 FROM latest l
		 INNER JOIN aggregated a ON a.problem_id = l.problem_id`,
		fixture.userID,
		fixture.contestID,
	)
	require.NoError(t, err)
	defer rows.Close()

	var planLines []string
	for rows.Next() {
		var line string
		require.NoError(t, rows.Scan(&line))
		planLines = append(planLines, line)
	}
	require.NoError(t, rows.Err())
	planText := strings.Join(planLines, "\n")
	require.True(t, regexp.MustCompile(`(?i)index scan|bitmap index scan`).MatchString(planText), planText)
}

func TestSQLRepositoryListUserProblemAttemptsP95Under300ms(t *testing.T) {
	ctx, pool := testDB(t)
	repo := contest_makeup.NewSQLRepository(pool)
	fixture := seedLargeContestFixture(t, ctx, pool, 20, 2000)

	const rounds = 60
	durations := make([]time.Duration, 0, rounds)
	for index := 0; index < rounds; index++ {
		start := time.Now()
		_, err := repo.ListUserProblemAttempts(ctx, fixture.userID, fixture.contestID)
		require.NoError(t, err)
		durations = append(durations, time.Since(start))
	}

	p95 := percentileDuration(durations, 0.95)
	require.Lessf(t, p95, 300*time.Millisecond, "p95=%s exceeds target under fixture 2k submissions / 20 problems", p95)
}

type contestFixture struct {
	userID    uuid.UUID
	contestID uuid.UUID
	problemA  uuid.UUID
	problemB  uuid.UUID
}

type largeContestFixture struct {
	userID    uuid.UUID
	contestID uuid.UUID
}

func seedContestFixture(t *testing.T, ctx context.Context, pool *pgxpool.Pool) contestFixture {
	t.Helper()

	suffix := strings.ReplaceAll(uuid.NewString(), "-", "")
	userID := uuid.New()
	contestID := uuid.New()
	problemA := uuid.New()
	problemB := uuid.New()
	problemC := uuid.New()

	_, err := pool.Exec(ctx,
		`INSERT INTO users (id, email, username, password_hash, role, status)
		 VALUES ($1, $2, $3, $4, $5, $6)`,
		userID,
		"integration-"+suffix+"@example.com",
		"user_"+suffix,
		"hash",
		"USER",
		"ACTIVE",
	)
	require.NoError(t, err)

	_, err = pool.Exec(ctx,
		`INSERT INTO contests (id, slug, title, status, starts_at, ends_at, blurb, rank_summary, remaining_label)
		 VALUES ($1, $2, $3, $4, $5, $6, '', '', '')`,
		contestID,
		"contest-"+suffix,
		"Contest "+suffix,
		"ENDED",
		time.Now().UTC().Add(-2*time.Hour),
		time.Now().UTC().Add(-time.Hour),
	)
	require.NoError(t, err)

	insertProblem := func(problemID uuid.UUID, slug string, title string, difficulty string) {
		_, innerErr := pool.Exec(ctx,
			`INSERT INTO problems (id, slug, title, difficulty, statement_md, input_md, output_md, constraints_md, time_limit_ms, memory_limit_kb, is_published)
			 VALUES ($1, $2, $3, $4, '', '', '', '', 1000, 262144, true)`,
			problemID,
			slug,
			title,
			difficulty,
		)
		require.NoError(t, innerErr)
	}

	insertProblem(problemA, "makeup-a-"+suffix, "Makeup A", "EASY")
	insertProblem(problemB, "makeup-b-"+suffix, "Makeup B", "MEDIUM")
	insertProblem(problemC, "makeup-c-"+suffix, "Makeup C", "HARD")

	_, err = pool.Exec(ctx,
		`INSERT INTO contest_problem_links (contest_id, problem_id, code, position)
		 VALUES
		   ($1, $2, 'A', 1),
		   ($1, $3, 'B', 2),
		   ($1, $4, 'C', 3)`,
		contestID,
		problemA,
		problemB,
		problemC,
	)
	require.NoError(t, err)

	_, err = pool.Exec(ctx,
		`INSERT INTO submissions (id, user_id, problem_id, contest_id, language, source_object, status, created_at)
		 VALUES
		   ($1, $2, $3, $4, 'CPP17', 'tmp/a-1.cpp', 'WRONG_ANSWER', now() - interval '20 minute'),
		   ($5, $2, $3, $4, 'CPP17', 'tmp/a-2.cpp', 'TIME_LIMIT_EXCEEDED', now() - interval '10 minute'),
		   ($6, $2, $7, $4, 'CPP17', 'tmp/b-1.cpp', 'ACCEPTED', now() - interval '5 minute')`,
		uuid.New(),
		userID,
		problemA,
		contestID,
		uuid.New(),
		uuid.New(),
		problemB,
	)
	require.NoError(t, err)

	return contestFixture{
		userID:    userID,
		contestID: contestID,
		problemA:  problemA,
		problemB:  problemB,
	}
}

func seedLargeContestFixture(t *testing.T, ctx context.Context, pool *pgxpool.Pool, problemCount int, submissionCount int) largeContestFixture {
	t.Helper()

	suffix := strings.ReplaceAll(uuid.NewString(), "-", "")
	userID := uuid.New()
	contestID := uuid.New()

	_, err := pool.Exec(ctx,
		`INSERT INTO users (id, email, username, password_hash, role, status)
		 VALUES ($1, $2, $3, $4, $5, $6)`,
		userID,
		"perf-"+suffix+"@example.com",
		"perf_"+suffix,
		"hash",
		"USER",
		"ACTIVE",
	)
	require.NoError(t, err)

	_, err = pool.Exec(ctx,
		`INSERT INTO contests (id, slug, title, status, starts_at, ends_at, blurb, rank_summary, remaining_label)
		 VALUES ($1, $2, $3, $4, $5, $6, '', '', '')`,
		contestID,
		"perf-contest-"+suffix,
		"Perf Contest "+suffix,
		"ENDED",
		time.Now().UTC().Add(-4*time.Hour),
		time.Now().UTC().Add(-2*time.Hour),
	)
	require.NoError(t, err)

	problemIDs := make([]uuid.UUID, 0, problemCount)
	for index := 0; index < problemCount; index++ {
		problemID := uuid.New()
		problemIDs = append(problemIDs, problemID)
		difficulty := "MEDIUM"
		switch index % 3 {
		case 0:
			difficulty = "EASY"
		case 2:
			difficulty = "HARD"
		}
		_, err = pool.Exec(ctx,
			`INSERT INTO problems (id, slug, title, difficulty, statement_md, input_md, output_md, constraints_md, time_limit_ms, memory_limit_kb, is_published)
			 VALUES ($1, $2, $3, $4, '', '', '', '', 1000, 262144, true)`,
			problemID,
			fmt.Sprintf("perf-problem-%s-%d", suffix, index+1),
			fmt.Sprintf("Perf Problem %d", index+1),
			difficulty,
		)
		require.NoError(t, err)

		_, err = pool.Exec(ctx,
			`INSERT INTO contest_problem_links (contest_id, problem_id, code, position)
			 VALUES ($1, $2, $3, $4)`,
			contestID,
			problemID,
			fmt.Sprintf("P%02d", index+1),
			index+1,
		)
		require.NoError(t, err)
	}

	statuses := []string{
		"WRONG_ANSWER",
		"TIME_LIMIT_EXCEEDED",
		"RUNTIME_ERROR",
		"MEMORY_LIMIT_EXCEEDED",
		"COMPILE_ERROR",
		"PENDING",
		"RUNNING",
		"ACCEPTED",
	}

	for index := 0; index < submissionCount; index++ {
		problemID := problemIDs[index%len(problemIDs)]
		status := statuses[index%len(statuses)]
		_, err = pool.Exec(ctx,
			`INSERT INTO submissions (id, user_id, problem_id, contest_id, language, source_object, status, created_at)
			 VALUES ($1, $2, $3, $4, 'CPP17', $5, $6::submission_status, $7)`,
			uuid.New(),
			userID,
			problemID,
			contestID,
			fmt.Sprintf("tmp/perf-%d.cpp", index+1),
			status,
			time.Now().UTC().Add(-time.Duration(submissionCount-index)*time.Second),
		)
		require.NoError(t, err)
	}

	return largeContestFixture{
		userID:    userID,
		contestID: contestID,
	}
}

func percentileDuration(values []time.Duration, percentile float64) time.Duration {
	if len(values) == 0 {
		return 0
	}
	sorted := append([]time.Duration(nil), values...)
	sort.Slice(sorted, func(left int, right int) bool {
		return sorted[left] < sorted[right]
	})
	position := int(math.Ceil(percentile*float64(len(sorted)))) - 1
	if position < 0 {
		position = 0
	}
	if position >= len(sorted) {
		position = len(sorted) - 1
	}
	return sorted[position]
}

func testDB(t *testing.T) (context.Context, *pgxpool.Pool) {
	t.Helper()

	databaseURL := os.Getenv("DATABASE_URL")
	if databaseURL == "" {
		databaseURL = "postgres://oj:oj@localhost:5432/oj?sslmode=disable"
	}

	ctx, cancel := context.WithTimeout(context.Background(), 8*time.Second)
	t.Cleanup(cancel)

	pool, err := pgxpool.New(ctx, databaseURL)
	require.NoError(t, err)
	if pingErr := pool.Ping(ctx); pingErr != nil {
		pool.Close()
		t.Skipf("skipping integration test because database is unavailable: %v", pingErr)
	}
	t.Cleanup(func() {
		pool.Close()
	})

	applyStorageReadModelMigration(t, ctx, pool)
	applyProblemRouteCodeMigration(t, ctx, pool)
	applyContestMakeupIndexMigration(t, ctx, pool)
	return ctx, pool
}

func applyStorageReadModelMigration(t *testing.T, ctx context.Context, pool *pgxpool.Pool) {
	t.Helper()

	migrationPath := filepath.Join("..", "..", "..", "..", "packages", "database", "migrations", "000003_storage_read_models.sql")
	content, err := os.ReadFile(migrationPath)
	require.NoError(t, err)

	parts := strings.Split(string(content), "-- +goose Down")
	require.Len(t, parts, 2)

	upSQL := strings.TrimSpace(strings.TrimPrefix(parts[0], "-- +goose Up"))
	_, err = pool.Exec(ctx, upSQL)
	if err != nil && !strings.Contains(err.Error(), "already exists") && !strings.Contains(err.Error(), "already a member") {
		require.NoError(t, err)
	}
}

func applyProblemRouteCodeMigration(t *testing.T, ctx context.Context, pool *pgxpool.Pool) {
	t.Helper()

	migrationPath := filepath.Join("..", "..", "..", "..", "packages", "database", "migrations", "000005_problem_public_route_codes.sql")
	content, err := os.ReadFile(migrationPath)
	require.NoError(t, err)

	parts := strings.Split(string(content), "-- +goose Down")
	require.Len(t, parts, 2)

	upSQL := strings.TrimSpace(strings.TrimPrefix(parts[0], "-- +goose Up"))
	_, err = pool.Exec(ctx, upSQL)
	if err != nil && !strings.Contains(err.Error(), "already exists") && !strings.Contains(err.Error(), "already a member") {
		require.NoError(t, err)
	}
}

func applyContestMakeupIndexMigration(t *testing.T, ctx context.Context, pool *pgxpool.Pool) {
	t.Helper()

	migrationPath := filepath.Join("..", "..", "..", "..", "packages", "database", "migrations", "000006_contest_makeup_indexes.sql")
	content, err := os.ReadFile(migrationPath)
	require.NoError(t, err)

	parts := strings.Split(string(content), "-- +goose Down")
	require.Len(t, parts, 2)

	upSQL := strings.TrimSpace(strings.TrimPrefix(parts[0], "-- +goose Up"))
	_, err = pool.Exec(ctx, upSQL)
	if err != nil && !strings.Contains(err.Error(), "already exists") {
		require.NoError(t, err)
	}
}
