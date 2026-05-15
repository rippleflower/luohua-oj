package judge

import (
	"context"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"
)

type SQLRepository struct {
	pool *pgxpool.Pool
}

func NewSQLRepository(pool *pgxpool.Pool) *SQLRepository {
	return &SQLRepository{pool: pool}
}

func (r *SQLRepository) GetSubmission(ctx context.Context, submissionID uuid.UUID) (Submission, error) {
	var submission Submission
	err := r.pool.QueryRow(
		ctx,
		`SELECT id, problem_id, language::text, source_object
		 FROM submissions
		 WHERE id = $1`,
		submissionID,
	).Scan(&submission.ID, &submission.ProblemID, &submission.Language, &submission.SourceObject)
	return submission, err
}

func (r *SQLRepository) ListTestCases(ctx context.Context, problemID uuid.UUID) ([]TestCase, error) {
	rows, err := r.pool.Query(
		ctx,
		`SELECT id, input_object, output_object
		 FROM test_cases
		 WHERE problem_id = $1
		 ORDER BY created_at ASC, id ASC`,
		problemID,
	)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var testCases []TestCase
	for rows.Next() {
		var testCase TestCase
		if err := rows.Scan(&testCase.ID, &testCase.InputObject, &testCase.OutputObject); err != nil {
			return nil, err
		}
		testCases = append(testCases, testCase)
	}
	return testCases, rows.Err()
}

func (r *SQLRepository) SaveResult(ctx context.Context, submissionID uuid.UUID, testCaseID uuid.UUID, result RunResult) error {
	_, err := r.pool.Exec(
		ctx,
		`INSERT INTO submission_results (
		   submission_id,
		   test_case_id,
		   status,
		   time_ms,
		   memory_kb,
		   output_snippet,
		   error_snippet
		 )
		 VALUES ($1, $2, $3::submission_status, NULLIF($4, 0), NULLIF($5, 0), NULLIF($6, ''), NULLIF($7, ''))`,
		submissionID,
		testCaseID,
		string(result.Status),
		result.TimeMs,
		result.MemoryKB,
		result.Output,
		result.Error,
	)
	return err
}

func (r *SQLRepository) UpdateStatus(ctx context.Context, submissionID uuid.UUID, status Status, compileOutput string) error {
	_, err := r.pool.Exec(
		ctx,
		`UPDATE submissions
		 SET status = $2::submission_status,
		     compile_output = NULLIF($3, ''),
		     judged_at = CASE WHEN $2::submission_status IN ('ACCEPTED', 'WRONG_ANSWER', 'TIME_LIMIT_EXCEEDED', 'MEMORY_LIMIT_EXCEEDED', 'RUNTIME_ERROR', 'COMPILE_ERROR', 'SYSTEM_ERROR') THEN now() ELSE judged_at END
		 WHERE id = $1`,
		submissionID,
		string(status),
		compileOutput,
	)
	return err
}
