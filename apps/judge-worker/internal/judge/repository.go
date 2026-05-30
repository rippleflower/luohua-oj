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
		`SELECT s.id, s.problem_id, s.language::text, s.source_object, p.time_limit_ms, p.memory_limit_kb
		 FROM submissions s
		 INNER JOIN problems p ON p.id = s.problem_id
		 WHERE s.id = $1`,
		submissionID,
	).Scan(&submission.ID, &submission.ProblemID, &submission.Language, &submission.SourceObject, &submission.TimeLimitMs, &submission.MemoryLimitKB)
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

func (r *SQLRepository) UpdateStatus(ctx context.Context, submissionID uuid.UUID, status Status, compileOutput string, maxTimeMs *int32, maxMemoryKB *int32) error {
	_, err := r.pool.Exec(
		ctx,
		`UPDATE submissions
		 SET status = $2::submission_status,
		     compile_output = NULLIF($3, ''),
		     max_time_ms = $4,
		     max_memory_kb = $5,
		     judged_at = CASE WHEN $2::submission_status IN ('ACCEPTED', 'WRONG_ANSWER', 'TIME_LIMIT_EXCEEDED', 'MEMORY_LIMIT_EXCEEDED', 'RUNTIME_ERROR', 'COMPILE_ERROR', 'SYSTEM_ERROR') THEN now() ELSE judged_at END
		 WHERE id = $1`,
		submissionID,
		string(status),
		compileOutput,
		maxTimeMs,
		maxMemoryKB,
	)
	if err != nil {
		return err
	}
	if err := r.refreshSubmissionSummary(ctx, submissionID); err != nil {
		return err
	}
	if err := r.refreshSubmissionDetail(ctx, submissionID); err != nil {
		return err
	}
	if err := r.refreshProblemStats(ctx, submissionID); err != nil {
		return err
	}
	if err := r.refreshProblemSummary(ctx, submissionID); err != nil {
		return err
	}
	if err := r.refreshUserStats(ctx, submissionID); err != nil {
		return err
	}
	if err := r.refreshUserProblemStatus(ctx, submissionID); err != nil {
		return err
	}
	if err := r.refreshUserProfileSnapshot(ctx, submissionID); err != nil {
		return err
	}
	return r.refreshUserRecentSubmissions(ctx, submissionID)
}

func (r *SQLRepository) refreshSubmissionSummary(ctx context.Context, submissionID uuid.UUID) error {
	_, err := r.pool.Exec(
		ctx,
		`INSERT INTO submission_summaries (
		   submission_id,
		   user_id,
		   username,
		   problem_id,
		   problem_json,
		   language,
		   source_object_key,
		   status,
		   created_at,
		   updated_at
		 )
		 SELECT
		   s.id,
		   s.user_id,
		   u.username,
		   s.problem_id,
		   jsonb_build_object('id', p.id, 'slug', p.slug, 'title', p.title),
		   s.language,
		   s.source_object,
		   s.status,
		   s.created_at,
		   now()
		 FROM submissions s
		 INNER JOIN users u ON u.id = s.user_id
		 INNER JOIN problems p ON p.id = s.problem_id
		 WHERE s.id = $1
		 ON CONFLICT (submission_id) DO UPDATE SET
		   user_id = EXCLUDED.user_id,
		   username = EXCLUDED.username,
		   problem_id = EXCLUDED.problem_id,
		   problem_json = EXCLUDED.problem_json,
		   language = EXCLUDED.language,
		   source_object_key = EXCLUDED.source_object_key,
		   status = EXCLUDED.status,
		   created_at = EXCLUDED.created_at,
		   updated_at = EXCLUDED.updated_at`,
		submissionID,
	)
	return err
}

func (r *SQLRepository) refreshSubmissionDetail(ctx context.Context, submissionID uuid.UUID) error {
	_, err := r.pool.Exec(
		ctx,
		`INSERT INTO submission_details (
		   submission_id,
		   user_id,
		   username,
		   problem_id,
		   problem_json,
		   language,
		   source_object_key,
		   status,
		   created_at,
		   compile_summary,
		   results_json,
		   artifact_availability,
		   updated_at
		 )
		 SELECT
		   s.id,
		   s.user_id,
		   u.username,
		   s.problem_id,
		   jsonb_build_object('id', p.id, 'slug', p.slug, 'title', p.title),
		   s.language,
		   s.source_object,
		   s.status,
		   s.created_at,
		   jsonb_build_object(
		     'compileOutput', COALESCE(s.compile_output, ''),
		     'maxTimeMs', s.max_time_ms,
		     'maxMemoryKb', s.max_memory_kb,
		     'judgedAt', s.judged_at
		   ),
		   COALESCE((
		     SELECT jsonb_agg(
		       jsonb_build_object(
		         'testCaseId', sr.test_case_id,
		         'status', sr.status,
		         'timeMs', sr.time_ms,
		         'memoryKb', sr.memory_kb,
		         'outputSnippet', sr.output_snippet,
		         'errorSnippet', sr.error_snippet
		       )
		       ORDER BY sr.id ASC
		     )
		     FROM submission_results sr
		     WHERE sr.submission_id = s.id
		   ), '[]'::jsonb),
		   jsonb_build_object(
		     'sourceObjectKey', s.source_object,
		     'artifacts', COALESCE((
		       SELECT jsonb_agg(
		         jsonb_build_object(
		           'artifactType', sa.artifact_type,
		           'objectKey', sa.object_key,
		           'contentType', sa.content_type,
		           'contentEncoding', sa.content_encoding,
		           'expiresAt', sa.expires_at
		         )
		         ORDER BY sa.created_at ASC, sa.id ASC
		       )
		       FROM submission_artifacts sa
		       WHERE sa.submission_id = s.id
		     ), '[]'::jsonb)
		   ),
		   now()
		 FROM submissions s
		 INNER JOIN users u ON u.id = s.user_id
		 INNER JOIN problems p ON p.id = s.problem_id
		 WHERE s.id = $1
		 ON CONFLICT (submission_id) DO UPDATE SET
		   user_id = EXCLUDED.user_id,
		   username = EXCLUDED.username,
		   problem_id = EXCLUDED.problem_id,
		   problem_json = EXCLUDED.problem_json,
		   language = EXCLUDED.language,
		   source_object_key = EXCLUDED.source_object_key,
		   status = EXCLUDED.status,
		   created_at = EXCLUDED.created_at,
		   compile_summary = EXCLUDED.compile_summary,
		   results_json = EXCLUDED.results_json,
		   artifact_availability = EXCLUDED.artifact_availability,
		   updated_at = EXCLUDED.updated_at`,
		submissionID,
	)
	return err
}

func (r *SQLRepository) refreshProblemStats(ctx context.Context, submissionID uuid.UUID) error {
	_, err := r.pool.Exec(
		ctx,
		`WITH selected_problem AS (
		   SELECT problem_id
		   FROM submissions
		   WHERE id = $1
		 ),
		 stats AS (
		   SELECT
		     sp.problem_id,
		     COUNT(s.*)::int AS submission_count,
		     COUNT(*) FILTER (WHERE s.status = 'ACCEPTED')::int AS accepted_count
		   FROM selected_problem sp
		   LEFT JOIN submissions s ON s.problem_id = sp.problem_id
		   GROUP BY sp.problem_id
		 )
		 INSERT INTO problem_stats (problem_id, submission_count, accepted_count, accepted_rate, updated_at)
		 SELECT
		   stats.problem_id,
		   stats.submission_count,
		   stats.accepted_count,
		   CASE
		     WHEN stats.submission_count = 0 THEN 0
		     ELSE ROUND((stats.accepted_count::numeric * 100) / stats.submission_count, 2)
		   END,
		   now()
		 FROM stats
		 ON CONFLICT (problem_id) DO UPDATE SET
		   submission_count = EXCLUDED.submission_count,
		   accepted_count = EXCLUDED.accepted_count,
		   accepted_rate = EXCLUDED.accepted_rate,
		   updated_at = EXCLUDED.updated_at`,
		submissionID,
	)
	return err
}

func (r *SQLRepository) refreshProblemSummary(ctx context.Context, submissionID uuid.UUID) error {
	_, err := r.pool.Exec(
		ctx,
		`WITH selected_problem AS (
		   SELECT problem_id
		   FROM submissions
		   WHERE id = $1
		 )
		 INSERT INTO problem_public_summaries (
		   problem_id,
		   slug,
		   title,
		   difficulty,
		   tags_json,
		   accepted_rate,
		   statistics_json,
		   updated_at
		 )
		 SELECT
		   p.id,
		   p.slug,
		   p.title,
		   p.difficulty,
		   COALESCE((
		     SELECT jsonb_agg(pt.slug ORDER BY pt.slug)
		     FROM problem_tag_links ptl
		     INNER JOIN problem_tags pt ON pt.id = ptl.tag_id
		     WHERE ptl.problem_id = p.id
		   ), '[]'::jsonb),
		   COALESCE(ps.accepted_rate, 0),
		   jsonb_build_object(
		     'submissionCount', COALESCE(ps.submission_count, 0),
		     'acceptedCount', COALESCE(ps.accepted_count, 0)
		   ),
		   now()
		 FROM selected_problem sp
		 INNER JOIN problems p ON p.id = sp.problem_id
		 LEFT JOIN problem_stats ps ON ps.problem_id = p.id
		 ON CONFLICT (problem_id) DO UPDATE SET
		   slug = EXCLUDED.slug,
		   title = EXCLUDED.title,
		   difficulty = EXCLUDED.difficulty,
		   tags_json = EXCLUDED.tags_json,
		   accepted_rate = EXCLUDED.accepted_rate,
		   statistics_json = EXCLUDED.statistics_json,
		   updated_at = EXCLUDED.updated_at`,
		submissionID,
	)
	return err
}

func (r *SQLRepository) refreshUserStats(ctx context.Context, submissionID uuid.UUID) error {
	_, err := r.pool.Exec(
		ctx,
		`WITH selected_user AS (
		   SELECT user_id
		   FROM submissions
		   WHERE id = $1
		 ),
		 activity AS (
		   SELECT
		     su.user_id,
		     COUNT(s.*)::int AS submission_count,
		     COUNT(*) FILTER (WHERE s.status = 'ACCEPTED')::int AS accepted_count,
		     MAX(s.created_at) AS last_active_at
		   FROM selected_user su
		   LEFT JOIN submissions s ON s.user_id = su.user_id
		   GROUP BY su.user_id
		 ),
		 solved AS (
		   SELECT
		     su.user_id,
		     COUNT(DISTINCT s.problem_id)::int AS solved_count
		   FROM selected_user su
		   LEFT JOIN submissions s ON s.user_id = su.user_id AND s.status = 'ACCEPTED'
		   GROUP BY su.user_id
		 )
		 INSERT INTO user_stats (user_id, solved_count, submission_count, accepted_count, last_active_at, updated_at)
		 SELECT
		   activity.user_id,
		   COALESCE(solved.solved_count, 0),
		   COALESCE(activity.submission_count, 0),
		   COALESCE(activity.accepted_count, 0),
		   activity.last_active_at,
		   now()
		 FROM activity
		 LEFT JOIN solved ON solved.user_id = activity.user_id
		 ON CONFLICT (user_id) DO UPDATE SET
		   solved_count = EXCLUDED.solved_count,
		   submission_count = EXCLUDED.submission_count,
		   accepted_count = EXCLUDED.accepted_count,
		   last_active_at = EXCLUDED.last_active_at,
		   updated_at = EXCLUDED.updated_at`,
		submissionID,
	)
	return err
}

func (r *SQLRepository) refreshUserProblemStatus(ctx context.Context, submissionID uuid.UUID) error {
	_, err := r.pool.Exec(
		ctx,
		`WITH selected_submission AS (
		   SELECT user_id, problem_id
		   FROM submissions
		   WHERE id = $1
		 ),
		 latest AS (
		   SELECT DISTINCT ON (s.user_id, s.problem_id)
		     s.user_id,
		     s.problem_id,
		     s.status AS latest_status
		   FROM submissions s
		   INNER JOIN selected_submission ss
		     ON ss.user_id = s.user_id
		    AND ss.problem_id = s.problem_id
		   ORDER BY s.user_id, s.problem_id, s.created_at DESC, s.id DESC
		 ),
		 solved AS (
		   SELECT
		     s.user_id,
		     s.problem_id,
		     MIN(s.created_at) AS first_accepted_at
		   FROM submissions s
		   INNER JOIN selected_submission ss
		     ON ss.user_id = s.user_id
		    AND ss.problem_id = s.problem_id
		   WHERE s.status = 'ACCEPTED'
		   GROUP BY s.user_id, s.problem_id
		 )
		 INSERT INTO user_problem_statuses (user_id, problem_id, best_status, latest_status, first_accepted_at, updated_at)
		 SELECT
		   latest.user_id,
		   latest.problem_id,
		   CASE
		     WHEN solved.first_accepted_at IS NOT NULL THEN 'ACCEPTED'::submission_status
		     ELSE latest.latest_status
		   END,
		   latest.latest_status,
		   solved.first_accepted_at,
		   now()
		 FROM latest
		 LEFT JOIN solved ON solved.user_id = latest.user_id AND solved.problem_id = latest.problem_id
		 ON CONFLICT (user_id, problem_id) DO UPDATE SET
		   best_status = EXCLUDED.best_status,
		   latest_status = EXCLUDED.latest_status,
		   first_accepted_at = EXCLUDED.first_accepted_at,
		   updated_at = EXCLUDED.updated_at`,
		submissionID,
	)
	return err
}

func (r *SQLRepository) refreshUserProfileSnapshot(ctx context.Context, submissionID uuid.UUID) error {
	_, err := r.pool.Exec(
		ctx,
		`WITH selected_user AS (
		   SELECT user_id
		   FROM submissions
		   WHERE id = $1
		 )
		 INSERT INTO user_profile_snapshot (user_id, summary_json, updated_at)
		 SELECT
		   u.id,
		   jsonb_build_object(
		     'userId', u.id,
		     'username', u.username,
		     'displayName', COALESCE(up.display_name, u.username),
		     'role', u.role,
		     'status', u.status,
		     'stats', jsonb_build_object(
		       'solvedCount', COALESCE(us.solved_count, 0),
		       'submissionCount', COALESCE(us.submission_count, 0),
		       'acceptedCount', COALESCE(us.accepted_count, 0)
		     )
		   ),
		   now()
		 FROM selected_user su
		 INNER JOIN users u ON u.id = su.user_id
		 LEFT JOIN user_profiles up ON up.user_id = u.id
		 LEFT JOIN user_stats us ON us.user_id = u.id
		 ON CONFLICT (user_id) DO UPDATE SET
		   summary_json = EXCLUDED.summary_json,
		   updated_at = EXCLUDED.updated_at`,
		submissionID,
	)
	return err
}

func (r *SQLRepository) refreshUserRecentSubmissions(ctx context.Context, submissionID uuid.UUID) error {
	_, err := r.pool.Exec(
		ctx,
		`WITH selected_user AS (
		   SELECT user_id
		   FROM submissions
		   WHERE id = $1
		 )
		 INSERT INTO user_recent_submission_snapshot (user_id, submissions_json, updated_at)
		 SELECT
		   su.user_id,
		   COALESCE((
		     SELECT jsonb_agg(item.payload ORDER BY item.created_at DESC)
		     FROM (
		       SELECT
		         ss.created_at,
		         jsonb_build_object(
		           'id', ss.submission_id,
		           'problem', ss.problem_json,
		           'language', ss.language,
		           'status', ss.status,
		           'sourceObjectKey', ss.source_object_key,
		           'createdAt', ss.created_at
		         ) AS payload
		       FROM submission_summaries ss
		       WHERE ss.user_id = su.user_id
		       ORDER BY ss.created_at DESC, ss.submission_id DESC
		       LIMIT 10
		     ) item
		   ), '[]'::jsonb),
		   now()
		 FROM selected_user su
		 ON CONFLICT (user_id) DO UPDATE SET
		   submissions_json = EXCLUDED.submissions_json,
		   updated_at = EXCLUDED.updated_at`,
		submissionID,
	)
	return err
}
