package problem

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"

	"github.com/example/oj3/apps/api/internal/auth"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/jackc/pgx/v5/pgxpool"
)

type SQLRepository struct {
	db *pgxpool.Pool
}

func NewSQLRepository(database *pgxpool.Pool) *SQLRepository {
	return &SQLRepository{db: database}
}

func (r *SQLRepository) ListProblems(ctx context.Context) ([]Summary, error) {
	rows, err := r.db.Query(
		ctx,
		`SELECT problem_id, slug, title, difficulty::text, tags_json, accepted_rate, updated_at
		 FROM problem_public_summaries
		 ORDER BY updated_at DESC, slug ASC`,
	)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var problems []Summary
	for rows.Next() {
		var (
			item     Summary
			tagsJSON []byte
		)
		if err := rows.Scan(&item.ID, &item.Slug, &item.Title, &item.Difficulty, &tagsJSON, &item.AcceptedRate, &item.UpdatedAt); err != nil {
			return nil, err
		}
		if err := json.Unmarshal(tagsJSON, &item.Tags); err != nil {
			return nil, err
		}
		problems = append(problems, item)
	}

	return problems, rows.Err()
}

func (r *SQLRepository) GetProblemBySlug(ctx context.Context, slug string) (Detail, error) {
	var item Detail
	err := r.db.QueryRow(
		ctx,
		`SELECT problem_id, slug, title, difficulty::text, statement_json, samples_json, limits_json, metadata_json, updated_at
		 FROM problem_public_details
		 WHERE slug = $1`,
		slug,
	).Scan(
		&item.ID,
		&item.Slug,
		&item.Title,
		&item.Difficulty,
		&item.StatementJSON,
		&item.SamplesJSON,
		&item.LimitsJSON,
		&item.MetadataJSON,
		&item.UpdatedAt,
	)
	return item, err
}

func (r *SQLRepository) CreateAdminProblem(ctx context.Context, actor auth.AuthenticatedUser, input CreateAdminInput) (AdminProblem, error) {
	tx, err := r.db.Begin(ctx)
	if err != nil {
		return AdminProblem{}, err
	}
	defer tx.Rollback(ctx)

	var problemID uuid.UUID
	if err := tx.QueryRow(
		ctx,
		`INSERT INTO problems (
		   slug,
		   title,
		   difficulty,
		   statement_md,
		   input_md,
		   output_md,
		   constraints_md,
		   time_limit_ms,
		   memory_limit_kb,
		   is_published
		 ) VALUES ($1, $2, $3::problem_difficulty, $4, $5, $6, $7, $8, $9, false)
		 RETURNING id`,
		input.Slug,
		input.Title,
		input.Difficulty,
		defaultStatementMarkdown,
		defaultInputMarkdown,
		defaultOutputMarkdown,
		defaultConstraintsMarkdown,
		input.TimeLimitMs,
		input.MemoryLimitKb,
	).Scan(&problemID); err != nil {
		return AdminProblem{}, mapProblemWriteError(err)
	}

	var versionID uuid.UUID
	if err := tx.QueryRow(
		ctx,
		`INSERT INTO problem_versions (
		   problem_id,
		   version_no,
		   manifest_object_key,
		   public_snapshot_version,
		   judge_bundle_version,
		   status
		 ) VALUES ($1, 1, $2, 1, 1, 'DRAFT')
		 RETURNING id`,
		problemID,
		fmt.Sprintf("problems/%s/versions/1/manifest.json", problemID),
	).Scan(&versionID); err != nil {
		return AdminProblem{}, err
	}

	if _, err := tx.Exec(
		ctx,
		`UPDATE problems
		 SET current_public_version_id = $2,
		     current_judge_version_id = $2,
		     updated_at = now()
		 WHERE id = $1`,
		problemID,
		versionID,
	); err != nil {
		return AdminProblem{}, err
	}

	if err := recordProblemAudit(ctx, tx, actor, "admin.problem.created", problemID.String(), map[string]any{
		"slug":          input.Slug,
		"title":         input.Title,
		"difficulty":    input.Difficulty,
		"timeLimitMs":   input.TimeLimitMs,
		"memoryLimitKb": input.MemoryLimitKb,
		"versionNo":     1,
	}, input.Reason, input.IP); err != nil {
		return AdminProblem{}, err
	}

	if err := tx.Commit(ctx); err != nil {
		return AdminProblem{}, err
	}
	return r.getAdminProblem(ctx, problemID)
}

func (r *SQLRepository) UpdateAdminProblem(ctx context.Context, actor auth.AuthenticatedUser, input UpdateAdminInput) (AdminProblem, error) {
	tx, err := r.db.Begin(ctx)
	if err != nil {
		return AdminProblem{}, err
	}
	defer tx.Rollback(ctx)

	tag, err := tx.Exec(
		ctx,
		`UPDATE problems
		 SET slug = $2,
		     title = $3,
		     difficulty = $4::problem_difficulty,
		     time_limit_ms = $5,
		     memory_limit_kb = $6,
		     updated_at = now()
		 WHERE id = $1`,
		input.ProblemID,
		input.Slug,
		input.Title,
		input.Difficulty,
		input.TimeLimitMs,
		input.MemoryLimitKb,
	)
	if err != nil {
		return AdminProblem{}, mapProblemWriteError(err)
	}
	if tag.RowsAffected() == 0 {
		return AdminProblem{}, errors.New("problem not found")
	}

	if err := recordProblemAudit(ctx, tx, actor, "admin.problem.updated", input.ProblemID.String(), map[string]any{
		"slug":          input.Slug,
		"title":         input.Title,
		"difficulty":    input.Difficulty,
		"timeLimitMs":   input.TimeLimitMs,
		"memoryLimitKb": input.MemoryLimitKb,
	}, input.Reason, input.IP); err != nil {
		return AdminProblem{}, err
	}

	if err := tx.Commit(ctx); err != nil {
		return AdminProblem{}, err
	}
	return r.getAdminProblem(ctx, input.ProblemID)
}

func (r *SQLRepository) PublishAdminProblem(ctx context.Context, actor auth.AuthenticatedUser, input PublishAdminInput) (AdminProblem, error) {
	tx, err := r.db.Begin(ctx)
	if err != nil {
		return AdminProblem{}, err
	}
	defer tx.Rollback(ctx)

	var versionID uuid.UUID
	if err := tx.QueryRow(
		ctx,
		`UPDATE problems
		 SET is_published = true,
		     updated_at = now()
		 WHERE id = $1
		 RETURNING current_public_version_id`,
		input.ProblemID,
	).Scan(&versionID); err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return AdminProblem{}, errors.New("problem not found")
		}
		return AdminProblem{}, err
	}

	if _, err := tx.Exec(
		ctx,
		`UPDATE problem_versions
		 SET status = 'PUBLISHED',
		     updated_at = now()
		 WHERE id = $1`,
		versionID,
	); err != nil {
		return AdminProblem{}, err
	}

	if err := r.refreshProblemSummaryByID(ctx, tx, input.ProblemID); err != nil {
		return AdminProblem{}, err
	}
	if err := r.refreshProblemDetailByID(ctx, tx, input.ProblemID); err != nil {
		return AdminProblem{}, err
	}

	if err := recordProblemAudit(ctx, tx, actor, "admin.problem.published", input.ProblemID.String(), map[string]any{
		"published": true,
		"versionId": versionID.String(),
	}, input.Reason, input.IP); err != nil {
		return AdminProblem{}, err
	}

	if err := tx.Commit(ctx); err != nil {
		return AdminProblem{}, err
	}
	return r.getAdminProblem(ctx, input.ProblemID)
}

func (r *SQLRepository) getAdminProblem(ctx context.Context, problemID uuid.UUID) (AdminProblem, error) {
	var item AdminProblem
	err := r.db.QueryRow(
		ctx,
		`SELECT
		   p.id,
		   p.slug,
		   p.title,
		   p.difficulty::text,
		   p.time_limit_ms,
		   p.memory_limit_kb,
		   COALESCE(pv.status, CASE WHEN p.is_published THEN 'PUBLISHED' ELSE 'DRAFT' END),
		   COALESCE(pv.version_no, 1),
		   p.is_published,
		   COALESCE(ps.submission_count, 0),
		   COALESCE(ps.accepted_rate::float8, 0),
		   p.updated_at
		 FROM problems p
		 LEFT JOIN problem_versions pv ON pv.problem_id = p.id AND pv.id = p.current_public_version_id
		 LEFT JOIN problem_stats ps ON ps.problem_id = p.id
		 WHERE p.id = $1`,
		problemID,
	).Scan(
		&item.ID,
		&item.Slug,
		&item.Title,
		&item.Difficulty,
		&item.TimeLimitMs,
		&item.MemoryLimitKb,
		&item.Status,
		&item.CurrentVersionNo,
		&item.IsPublished,
		&item.SubmissionCount,
		&item.AcceptedRate,
		&item.UpdatedAt,
	)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return AdminProblem{}, errors.New("problem not found")
		}
		return AdminProblem{}, err
	}
	return item, nil
}

type problemExec interface {
	Exec(context.Context, string, ...any) (pgconn.CommandTag, error)
}

func (r *SQLRepository) refreshProblemSummaryByID(ctx context.Context, exec problemExec, problemID uuid.UUID) error {
	_, err := exec.Exec(
		ctx,
		`INSERT INTO problem_public_summaries (
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
		 FROM problems p
		 LEFT JOIN problem_stats ps ON ps.problem_id = p.id
		 WHERE p.id = $1
		 ON CONFLICT (problem_id) DO UPDATE SET
		   slug = EXCLUDED.slug,
		   title = EXCLUDED.title,
		   difficulty = EXCLUDED.difficulty,
		   tags_json = EXCLUDED.tags_json,
		   accepted_rate = EXCLUDED.accepted_rate,
		   statistics_json = EXCLUDED.statistics_json,
		   updated_at = EXCLUDED.updated_at`,
		problemID,
	)
	return err
}

func (r *SQLRepository) refreshProblemDetailByID(ctx context.Context, exec problemExec, problemID uuid.UUID) error {
	_, err := exec.Exec(
		ctx,
		`INSERT INTO problem_public_details (
		   problem_id,
		   slug,
		   title,
		   difficulty,
		   statement_json,
		   samples_json,
		   limits_json,
		   metadata_json,
		   updated_at
		 )
		 SELECT
		   p.id,
		   p.slug,
		   p.title,
		   p.difficulty,
		   jsonb_build_array(
		     jsonb_build_object('kind', 'markdown', 'section', 'statement', 'content', p.statement_md),
		     jsonb_build_object('kind', 'markdown', 'section', 'input', 'content', p.input_md),
		     jsonb_build_object('kind', 'markdown', 'section', 'output', 'content', p.output_md),
		     jsonb_build_object('kind', 'markdown', 'section', 'constraints', 'content', p.constraints_md)
		   ),
		   COALESCE((
		     SELECT jsonb_agg(
		       jsonb_build_object(
		         'inputObjectKey', tc.input_object,
		         'outputObjectKey', tc.output_object,
		         'weight', tc.weight
		       )
		       ORDER BY tc.created_at ASC, tc.id ASC
		     )
		     FROM test_cases tc
		     WHERE tc.problem_id = p.id AND tc.is_sample = true
		   ), '[]'::jsonb),
		   jsonb_build_object('timeLimitMs', p.time_limit_ms, 'memoryLimitKb', p.memory_limit_kb),
		   jsonb_build_object(
		     'published', p.is_published,
		     'publicVersionId', p.current_public_version_id,
		     'judgeVersionId', p.current_judge_version_id
		   ),
		   now()
		 FROM problems p
		 WHERE p.id = $1
		 ON CONFLICT (problem_id) DO UPDATE SET
		   slug = EXCLUDED.slug,
		   title = EXCLUDED.title,
		   difficulty = EXCLUDED.difficulty,
		   statement_json = EXCLUDED.statement_json,
		   samples_json = EXCLUDED.samples_json,
		   limits_json = EXCLUDED.limits_json,
		   metadata_json = EXCLUDED.metadata_json,
		   updated_at = EXCLUDED.updated_at`,
		problemID,
	)
	return err
}

func mapProblemWriteError(err error) error {
	var pgErr *pgconn.PgError
	if errors.As(err, &pgErr) && pgErr.Code == "23505" {
		return errors.New("slug already exists")
	}
	return err
}

func recordProblemAudit(ctx context.Context, exec problemExec, actor auth.AuthenticatedUser, action string, targetID string, diff map[string]any, reason string, ip string) error {
	diffJSON, err := json.Marshal(diff)
	if err != nil {
		return err
	}
	_, err = exec.Exec(
		ctx,
		`INSERT INTO audit_logs (actor_user_id, actor_role, action, target_type, target_id, diff_json, reason, ip)
		 VALUES ($1, $2, $3, 'problem', $4, $5, $6, $7)`,
		actor.ID,
		actor.Role,
		action,
		targetID,
		diffJSON,
		reason,
		ip,
	)
	return err
}

const (
	defaultStatementMarkdown   = "题面待补充。"
	defaultInputMarkdown       = "输入说明待补充。"
	defaultOutputMarkdown      = "输出说明待补充。"
	defaultConstraintsMarkdown = "数据范围待补充。"
)
