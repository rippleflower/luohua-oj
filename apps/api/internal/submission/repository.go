package submission

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"time"

	"github.com/example/oj3/apps/api/internal/auth"
	db "github.com/example/oj3/apps/api/internal/db/generated"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgtype"
)

type SQLRepository struct {
	db      db.DBTX
	queries db.Querier
}

func NewSQLRepository(database db.DBTX) *SQLRepository {
	return &SQLRepository{
		db:      database,
		queries: db.New(database),
	}
}

func (r *SQLRepository) CreateSubmission(ctx context.Context, params CreateParams) (Submission, error) {
	created, err := r.queries.CreateSubmission(ctx, db.CreateSubmissionParams{
		ID:           uuidToPG(params.ID),
		UserID:       uuidToPG(params.UserID),
		ProblemID:    uuidToPG(params.ProblemID),
		Language:     db.Language(params.Language),
		SourceObject: params.SourceObjectKey,
	})
	if err != nil {
		return Submission{}, err
	}

	return submissionSummaryFromDB(created)
}

func (r *SQLRepository) GetSubmission(ctx context.Context, submissionID uuid.UUID) (SubmissionDetail, error) {
	var (
		summary      Submission
		problemJSON  []byte
		compileJSON  []byte
		resultsJSON  []byte
		artifactJSON []byte
	)

	err := r.db.QueryRow(
		ctx,
		`SELECT
		   submission_id,
		   user_id,
		   problem_id,
		   problem_json,
		   language::text,
		   source_object_key,
		   status::text,
		   created_at,
		   compile_summary,
		   results_json,
		   artifact_availability
		 FROM submission_details
		 WHERE submission_id = $1`,
		submissionID,
	).Scan(
		&summary.ID,
		&summary.UserID,
		&summary.ProblemID,
		&problemJSON,
		&summary.Language,
		&summary.SourceObjectKey,
		&summary.Status,
		&summary.CreatedAt,
		&compileJSON,
		&resultsJSON,
		&artifactJSON,
	)
	if err != nil {
		return SubmissionDetail{}, err
	}

	problem, err := parseProblemSummaryJSON(problemJSON)
	if err != nil {
		return SubmissionDetail{}, fmt.Errorf("parse problem summary: %w", err)
	}
	summary.Problem = problem

	compileSummary, err := parseCompileSummaryJSON(compileJSON)
	if err != nil {
		return SubmissionDetail{}, fmt.Errorf("parse compile summary: %w", err)
	}
	results, err := parseResultsJSON(resultsJSON)
	if err != nil {
		return SubmissionDetail{}, fmt.Errorf("parse results: %w", err)
	}
	artifactAvailability, err := parseArtifactAvailabilityJSON(artifactJSON)
	if err != nil {
		return SubmissionDetail{}, fmt.Errorf("parse artifacts: %w", err)
	}

	return SubmissionDetail{
		Submission:           summary,
		CompileSummary:       compileSummary,
		Results:              results,
		ArtifactAvailability: artifactAvailability,
	}, nil
}

func (r *SQLRepository) ListSubmissionsByUsername(ctx context.Context, params ListByUsernameParams) (SubmissionPage, error) {
	offset := (params.Page - 1) * params.PageSize
	rows, err := r.db.Query(
		ctx,
		`SELECT
		   submission_id,
		   user_id,
		   problem_id,
		   problem_json,
		   language::text,
		   source_object_key,
		   status::text,
		   created_at,
		   COUNT(*) OVER()
		 FROM submission_summaries
		 WHERE username = $1
		 ORDER BY created_at DESC, submission_id DESC
		 LIMIT $2 OFFSET $3`,
		params.Username,
		params.PageSize,
		offset,
	)
	if err != nil {
		return SubmissionPage{}, err
	}
	defer rows.Close()

	var items []Submission
	total := 0
	for rows.Next() {
		var (
			item        Submission
			problemJSON []byte
			totalCount  int64
		)
		if err := rows.Scan(
			&item.ID,
			&item.UserID,
			&item.ProblemID,
			&problemJSON,
			&item.Language,
			&item.SourceObjectKey,
			&item.Status,
			&item.CreatedAt,
			&totalCount,
		); err != nil {
			return SubmissionPage{}, err
		}
		problem, err := parseProblemSummaryJSON(problemJSON)
		if err != nil {
			return SubmissionPage{}, fmt.Errorf("parse problem summary: %w", err)
		}
		item.Problem = problem
		items = append(items, item)
		total = int(totalCount)
	}
	if err := rows.Err(); err != nil {
		return SubmissionPage{}, err
	}

	return SubmissionPage{
		Items:    items,
		Total:    total,
		Page:     params.Page,
		PageSize: params.PageSize,
	}, nil
}

func (r *SQLRepository) RefreshSubmissionViews(ctx context.Context, submissionID uuid.UUID) error {
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
	if err := r.refreshUserRecentSubmissions(ctx, submissionID); err != nil {
		return err
	}
	return nil
}

func (r *SQLRepository) GetRejudgeTarget(ctx context.Context, submissionID uuid.UUID) (RejudgeTarget, error) {
	var (
		target                 RejudgeTarget
		contestID              *uuid.UUID
		currentProblemVersion  *uuid.UUID
		snapshotProblemVersion *uuid.UUID
	)

	err := r.db.QueryRow(
		ctx,
		`SELECT
		   s.id,
		   s.contest_id,
		   s.result_snapshot_version,
		   p.current_judge_version_id,
		   cps.problem_version_id
		 FROM submissions s
		 INNER JOIN problems p ON p.id = s.problem_id
		 LEFT JOIN contest_snapshots cs
		   ON cs.contest_id = s.contest_id
		  AND cs.snapshot_no = s.result_snapshot_version
		 LEFT JOIN contest_problem_snapshots cps
		   ON cps.snapshot_id = cs.id
		  AND cps.problem_id = s.problem_id
		 WHERE s.id = $1`,
		submissionID,
	).Scan(
		&target.SubmissionID,
		&contestID,
		&target.ResultSnapshotVersion,
		&currentProblemVersion,
		&snapshotProblemVersion,
	)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return RejudgeTarget{}, errors.New("submission not found")
		}
		return RejudgeTarget{}, err
	}

	target.ContestID = contestID
	target.CurrentProblemVersionID = currentProblemVersion
	target.SnapshotProblemVersionID = snapshotProblemVersion
	return target, nil
}

func (r *SQLRepository) ResetSubmissionForRejudge(ctx context.Context, actor auth.AuthenticatedUser, input RejudgeAdminInput, resultSnapshotVersion int, problemVersionID *uuid.UUID) error {
	tx, err := beginTx(ctx, r.db)
	if err != nil {
		return err
	}
	defer tx.Rollback(ctx)

	tag, err := tx.Exec(
		ctx,
		`UPDATE submissions
		 SET status = 'PENDING',
		     score = 0,
		     compile_output = NULL,
		     max_time_ms = NULL,
		     max_memory_kb = NULL,
		     judged_at = NULL,
		     result_snapshot_version = $2
		 WHERE id = $1`,
		input.SubmissionID,
		resultSnapshotVersion,
	)
	if err != nil {
		return err
	}
	if tag.RowsAffected() == 0 {
		return errors.New("submission not found")
	}

	if _, err := tx.Exec(ctx, `DELETE FROM submission_results WHERE submission_id = $1`, input.SubmissionID); err != nil {
		return err
	}
	if _, err := tx.Exec(ctx, `DELETE FROM submission_artifacts WHERE submission_id = $1`, input.SubmissionID); err != nil {
		return err
	}

	repo := &SQLRepository{db: tx, queries: db.New(tx)}
	if err := repo.RefreshSubmissionViews(ctx, input.SubmissionID); err != nil {
		return err
	}
	diffJSON, err := json.Marshal(map[string]any{
		"resultSnapshotVersion": resultSnapshotVersion,
		"problemVersionId":      nullableUUIDString(problemVersionID),
	})
	if err != nil {
		return err
	}
	if _, err := tx.Exec(
		ctx,
		`INSERT INTO audit_logs (actor_user_id, actor_role, action, target_type, target_id, diff_json, reason, ip)
		 VALUES ($1, $2, 'admin.submission.rejudged', 'submission', $3, $4, $5, $6)`,
		actor.ID,
		actor.Role,
		input.SubmissionID.String(),
		diffJSON,
		input.Reason,
		input.IP,
	); err != nil {
		return err
	}
	return tx.Commit(ctx)
}

func (r *SQLRepository) refreshSubmissionSummary(ctx context.Context, submissionID uuid.UUID) error {
	_, err := r.db.Exec(
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
	_, err := r.db.Exec(
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
	_, err := r.db.Exec(
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
	_, err := r.db.Exec(
		ctx,
		`WITH selected_problem AS (
		   SELECT problem_id
		   FROM submissions
		   WHERE id = $1
		 )
		 INSERT INTO problem_public_summaries (
		   problem_id,
		   problem_no,
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
		   p.problem_no,
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
		   problem_no = EXCLUDED.problem_no,
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
	_, err := r.db.Exec(
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
	_, err := r.db.Exec(
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
	_, err := r.db.Exec(
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
	_, err := r.db.Exec(
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

func submissionSummaryFromDB(model db.Submission) (Submission, error) {
	id, err := uuidFromPG(model.ID)
	if err != nil {
		return Submission{}, fmt.Errorf("submission id: %w", err)
	}
	userID, err := uuidFromPG(model.UserID)
	if err != nil {
		return Submission{}, fmt.Errorf("user id: %w", err)
	}
	problemID, err := uuidFromPG(model.ProblemID)
	if err != nil {
		return Submission{}, fmt.Errorf("problem id: %w", err)
	}

	return Submission{
		ID:              id,
		UserID:          userID,
		ProblemID:       problemID,
		Language:        string(model.Language),
		SourceObjectKey: model.SourceObject,
		Status:          Status(model.Status),
		CreatedAt:       timestampFromPG(model.CreatedAt),
	}, nil
}

func parseProblemSummaryJSON(payload []byte) (*ProblemSummary, error) {
	var raw struct {
		ID    string `json:"id"`
		Slug  string `json:"slug"`
		Title string `json:"title"`
	}
	if err := json.Unmarshal(payload, &raw); err != nil {
		return nil, err
	}
	problemID, err := uuid.Parse(raw.ID)
	if err != nil {
		return nil, err
	}
	return &ProblemSummary{
		ID:    problemID,
		Slug:  raw.Slug,
		Title: raw.Title,
	}, nil
}

func parseCompileSummaryJSON(payload []byte) (CompileSummary, error) {
	var raw struct {
		CompileOutput string  `json:"compileOutput"`
		MaxTimeMs     *int32  `json:"maxTimeMs"`
		MaxMemoryKb   *int32  `json:"maxMemoryKb"`
		JudgedAt      *string `json:"judgedAt"`
	}
	if err := json.Unmarshal(payload, &raw); err != nil {
		return CompileSummary{}, err
	}
	summary := CompileSummary{
		CompileOutput: raw.CompileOutput,
		MaxTimeMs:     raw.MaxTimeMs,
		MaxMemoryKb:   raw.MaxMemoryKb,
	}
	if raw.JudgedAt != nil && *raw.JudgedAt != "" {
		parsed, err := time.Parse(time.RFC3339Nano, *raw.JudgedAt)
		if err != nil {
			return CompileSummary{}, err
		}
		summary.JudgedAt = &parsed
	}
	return summary, nil
}

func parseResultsJSON(payload []byte) ([]ResultSummary, error) {
	var raw []struct {
		TestCaseID    string `json:"testCaseId"`
		Status        string `json:"status"`
		TimeMs        *int32 `json:"timeMs"`
		MemoryKb      *int32 `json:"memoryKb"`
		OutputSnippet string `json:"outputSnippet"`
		ErrorSnippet  string `json:"errorSnippet"`
	}
	if err := json.Unmarshal(payload, &raw); err != nil {
		return nil, err
	}
	results := make([]ResultSummary, 0, len(raw))
	for _, item := range raw {
		testCaseID, err := uuid.Parse(item.TestCaseID)
		if err != nil {
			return nil, err
		}
		results = append(results, ResultSummary{
			TestCaseID:    testCaseID,
			Status:        Status(item.Status),
			TimeMs:        item.TimeMs,
			MemoryKb:      item.MemoryKb,
			OutputSnippet: item.OutputSnippet,
			ErrorSnippet:  item.ErrorSnippet,
		})
	}
	return results, nil
}

func parseArtifactAvailabilityJSON(payload []byte) (ArtifactAvailability, error) {
	var raw struct {
		SourceObjectKey string `json:"sourceObjectKey"`
		Artifacts       []struct {
			ArtifactType    string  `json:"artifactType"`
			ObjectKey       string  `json:"objectKey"`
			ContentType     string  `json:"contentType"`
			ContentEncoding string  `json:"contentEncoding"`
			ExpiresAt       *string `json:"expiresAt"`
		} `json:"artifacts"`
	}
	if err := json.Unmarshal(payload, &raw); err != nil {
		return ArtifactAvailability{}, err
	}
	availability := ArtifactAvailability{
		SourceObjectKey: raw.SourceObjectKey,
		Artifacts:       make([]ArtifactSummary, 0, len(raw.Artifacts)),
	}
	for _, item := range raw.Artifacts {
		artifact := ArtifactSummary{
			ArtifactType:    item.ArtifactType,
			ObjectKey:       item.ObjectKey,
			ContentType:     item.ContentType,
			ContentEncoding: item.ContentEncoding,
		}
		if item.ExpiresAt != nil && *item.ExpiresAt != "" {
			parsed, err := time.Parse(time.RFC3339Nano, *item.ExpiresAt)
			if err != nil {
				return ArtifactAvailability{}, err
			}
			artifact.ExpiresAt = &parsed
		}
		availability.Artifacts = append(availability.Artifacts, artifact)
	}
	return availability, nil
}

func uuidToPG(id uuid.UUID) pgtype.UUID {
	return pgtype.UUID{Bytes: id, Valid: true}
}

func uuidFromPG(id pgtype.UUID) (uuid.UUID, error) {
	if !id.Valid {
		return uuid.Nil, fmt.Errorf("uuid is null")
	}
	return uuid.UUID(id.Bytes), nil
}

func timestampFromPG(value pgtype.Timestamptz) time.Time {
	if !value.Valid {
		return time.Time{}
	}
	return value.Time.UTC()
}

func nullableUUIDString(value *uuid.UUID) any {
	if value == nil {
		return nil
	}
	return value.String()
}

type txStarter interface {
	Begin(ctx context.Context) (pgx.Tx, error)
}

func beginTx(ctx context.Context, database db.DBTX) (pgx.Tx, error) {
	starter, ok := database.(txStarter)
	if !ok {
		return nil, errors.New("database does not support transactions")
	}
	return starter.Begin(ctx)
}

var _ pgx.Row
