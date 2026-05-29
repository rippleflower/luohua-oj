package problem

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"

	"github.com/example/oj3/apps/api/internal/auth"
	"github.com/example/oj3/apps/api/internal/source"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/jackc/pgx/v5/pgxpool"
)

type SQLRepository struct {
	db    *pgxpool.Pool
	store source.LocalStore
}

func NewSQLRepository(database *pgxpool.Pool, store source.LocalStore) *SQLRepository {
	return &SQLRepository{db: database, store: store}
}

func (r *SQLRepository) ListProblems(ctx context.Context) ([]Summary, error) {
	rows, err := r.db.Query(
		ctx,
		`SELECT problem_id, problem_no, slug, title, difficulty::text, tags_json, accepted_rate, updated_at
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
		if err := rows.Scan(&item.ID, &item.ProblemNo, &item.Slug, &item.Title, &item.Difficulty, &tagsJSON, &item.AcceptedRate, &item.UpdatedAt); err != nil {
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
		`SELECT problem_id, problem_no, slug, title, difficulty::text, statement_json, samples_json, limits_json, metadata_json, updated_at
		 FROM problem_public_details
		 WHERE slug = $1`,
		slug,
	).Scan(
		&item.ID,
		&item.ProblemNo,
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

func (r *SQLRepository) GetProblemByNumber(ctx context.Context, problemNo int64) (Detail, error) {
	var item Detail
	err := r.db.QueryRow(
		ctx,
		`SELECT problem_id, problem_no, slug, title, difficulty::text, statement_json, samples_json, limits_json, metadata_json, updated_at
		 FROM problem_public_details
		 WHERE problem_no = $1`,
		problemNo,
	).Scan(
		&item.ID,
		&item.ProblemNo,
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

func (r *SQLRepository) GetAdminProblemDetail(ctx context.Context, problemID uuid.UUID) (AdminProblemDetail, error) {
	item, err := r.getAdminProblem(ctx, problemID)
	if err != nil {
		return AdminProblemDetail{}, err
	}

	statements, samples, tags, err := r.loadProblemContent(ctx, r.db, problemID)
	if err != nil {
		return AdminProblemDetail{}, err
	}

	return AdminProblemDetail{
		AdminProblem:  item,
		StatementJSON: statements,
		Samples:       samples,
		Tags:          tags,
	}, nil
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

func (r *SQLRepository) UpdateAdminProblemContent(ctx context.Context, actor auth.AuthenticatedUser, input UpdateAdminContentInput) (AdminProblemDetail, error) {
	storedSamples, err := r.writeSampleObjects(ctx, input.ProblemID, input.Samples)
	if err != nil {
		return AdminProblemDetail{}, err
	}

	tx, err := r.db.Begin(ctx)
	if err != nil {
		return AdminProblemDetail{}, err
	}
	defer tx.Rollback(ctx)

	statement := statementContent(input.StatementJSON, "statement")
	inputSection := statementContent(input.StatementJSON, "input")
	outputSection := statementContent(input.StatementJSON, "output")
	constraints := statementContent(input.StatementJSON, "constraints")

	tag, err := tx.Exec(
		ctx,
		`UPDATE problems
		 SET statement_md = $2,
		     input_md = $3,
		     output_md = $4,
		     constraints_md = $5,
		     updated_at = now()
		 WHERE id = $1`,
		input.ProblemID,
		statement,
		inputSection,
		outputSection,
		constraints,
	)
	if err != nil {
		return AdminProblemDetail{}, err
	}
	if tag.RowsAffected() == 0 {
		return AdminProblemDetail{}, errors.New("problem not found")
	}

	if _, err := tx.Exec(ctx, `DELETE FROM problem_tag_links WHERE problem_id = $1`, input.ProblemID); err != nil {
		return AdminProblemDetail{}, err
	}
	for _, problemTag := range input.Tags {
		if _, err := tx.Exec(
			ctx,
			`INSERT INTO problem_tags (slug, label)
			 VALUES ($1, $2)
			 ON CONFLICT (slug) DO NOTHING`,
			problemTag,
			problemTag,
		); err != nil {
			return AdminProblemDetail{}, err
		}
		if _, err := tx.Exec(
			ctx,
			`INSERT INTO problem_tag_links (problem_id, tag_id)
			 SELECT $1, id
			 FROM problem_tags
			 WHERE slug = $2
			 ON CONFLICT (problem_id, tag_id) DO NOTHING`,
			input.ProblemID,
			problemTag,
		); err != nil {
			return AdminProblemDetail{}, err
		}
	}

	if _, err := tx.Exec(ctx, `DELETE FROM test_cases WHERE problem_id = $1 AND is_sample = true`, input.ProblemID); err != nil {
		return AdminProblemDetail{}, err
	}
	for _, sample := range storedSamples {
		if _, err := tx.Exec(
			ctx,
			`INSERT INTO test_cases (problem_id, input_object, output_object, is_sample, weight)
			 VALUES ($1, $2, $3, true, $4)`,
			input.ProblemID,
			sample.InputObjectKey,
			sample.OutputObjectKey,
			sample.Weight,
		); err != nil {
			return AdminProblemDetail{}, err
		}
	}

	if err := recordProblemAudit(ctx, tx, actor, "admin.problem.content.updated", input.ProblemID.String(), map[string]any{
		"tags":         input.Tags,
		"sampleCount":  len(input.Samples),
		"statementLen": len(statement),
	}, input.Reason, input.IP); err != nil {
		return AdminProblemDetail{}, err
	}

	if err := tx.Commit(ctx); err != nil {
		return AdminProblemDetail{}, err
	}
	return r.GetAdminProblemDetail(ctx, input.ProblemID)
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
		   p.problem_no,
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
		&item.ProblemNo,
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

type problemQueries interface {
	Exec(context.Context, string, ...any) (pgconn.CommandTag, error)
	Query(context.Context, string, ...any) (pgx.Rows, error)
	QueryRow(context.Context, string, ...any) pgx.Row
}

func (r *SQLRepository) refreshProblemSummaryByID(ctx context.Context, exec problemQueries, problemID uuid.UUID) error {
	_, err := exec.Exec(
		ctx,
		`INSERT INTO problem_public_summaries (
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
		 FROM problems p
		 LEFT JOIN problem_stats ps ON ps.problem_id = p.id
		 WHERE p.id = $1
		 ON CONFLICT (problem_id) DO UPDATE SET
		   problem_no = EXCLUDED.problem_no,
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

func (r *SQLRepository) refreshProblemDetailByID(ctx context.Context, exec problemQueries, problemID uuid.UUID) error {
	state, err := r.loadProblemPublicState(ctx, exec, problemID)
	if err != nil {
		return err
	}

	statementJSON, err := json.Marshal(state.StatementJSON)
	if err != nil {
		return err
	}
	samplesJSON, err := json.Marshal(state.Samples)
	if err != nil {
		return err
	}
	limitsJSON, err := json.Marshal(map[string]any{
		"timeLimitMs":   state.TimeLimitMs,
		"memoryLimitKb": state.MemoryLimitKB,
	})
	if err != nil {
		return err
	}
	metadataJSON, err := json.Marshal(map[string]any{
		"published":       state.IsPublished,
		"publicVersionId": nullableUUIDStringValue(state.PublicVersionID),
		"judgeVersionId":  nullableUUIDStringValue(state.JudgeVersionID),
		"tags":            state.Tags,
		"acceptedRate":    state.AcceptedRate,
	})
	if err != nil {
		return err
	}

	_, err = exec.Exec(
		ctx,
		`INSERT INTO problem_public_details (
		   problem_id,
		   problem_no,
		   slug,
		   title,
		   difficulty,
		   statement_json,
		   samples_json,
		   limits_json,
		   metadata_json,
		   updated_at
		 ) VALUES ($1, $2, $3, $4, $5::problem_difficulty, $6::jsonb, $7::jsonb, $8::jsonb, $9::jsonb, now())
		 ON CONFLICT (problem_id) DO UPDATE SET
		   problem_no = EXCLUDED.problem_no,
		   slug = EXCLUDED.slug,
		   title = EXCLUDED.title,
		   difficulty = EXCLUDED.difficulty,
		   statement_json = EXCLUDED.statement_json,
		   samples_json = EXCLUDED.samples_json,
		   limits_json = EXCLUDED.limits_json,
		   metadata_json = EXCLUDED.metadata_json,
		   updated_at = EXCLUDED.updated_at`,
		state.ID,
		state.ProblemNo,
		state.Slug,
		state.Title,
		state.Difficulty,
		statementJSON,
		samplesJSON,
		limitsJSON,
		metadataJSON,
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

func recordProblemAudit(ctx context.Context, exec problemQueries, actor auth.AuthenticatedUser, action string, targetID string, diff map[string]any, reason string, ip string) error {
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

type storedSample struct {
	InputObjectKey  string
	OutputObjectKey string
	Weight          int
}

type publicProblemState struct {
	ID              uuid.UUID
	ProblemNo       int64
	Slug            string
	Title           string
	Difficulty      string
	TimeLimitMs     int
	MemoryLimitKB   int
	IsPublished     bool
	PublicVersionID *uuid.UUID
	JudgeVersionID  *uuid.UUID
	AcceptedRate    float64
	StatementJSON   []StatementSection
	Samples         []Sample
	Tags            []string
}

func (r *SQLRepository) loadProblemContent(ctx context.Context, queries problemQueries, problemID uuid.UUID) ([]StatementSection, []Sample, []string, error) {
	var (
		statementMD   string
		inputMD       string
		outputMD      string
		constraintsMD string
	)
	if err := queries.QueryRow(
		ctx,
		`SELECT statement_md, input_md, output_md, constraints_md
		 FROM problems
		 WHERE id = $1`,
		problemID,
	).Scan(&statementMD, &inputMD, &outputMD, &constraintsMD); err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, nil, nil, errors.New("problem not found")
		}
		return nil, nil, nil, err
	}

	rows, err := queries.Query(
		ctx,
		`SELECT input_object, output_object, weight
		 FROM test_cases
		 WHERE problem_id = $1 AND is_sample = true
		 ORDER BY created_at ASC, id ASC`,
		problemID,
	)
	if err != nil {
		return nil, nil, nil, err
	}
	defer rows.Close()

	samples := make([]Sample, 0)
	for rows.Next() {
		var sample storedSample
		if err := rows.Scan(&sample.InputObjectKey, &sample.OutputObjectKey, &sample.Weight); err != nil {
			return nil, nil, nil, err
		}
		inputContent, err := r.store.ReadObject(sample.InputObjectKey)
		if err != nil {
			return nil, nil, nil, err
		}
		outputContent, err := r.store.ReadObject(sample.OutputObjectKey)
		if err != nil {
			return nil, nil, nil, err
		}
		samples = append(samples, Sample{
			Input:  string(inputContent),
			Output: string(outputContent),
			Weight: sample.Weight,
		})
	}
	if err := rows.Err(); err != nil {
		return nil, nil, nil, err
	}

	tagRows, err := queries.Query(
		ctx,
		`SELECT pt.slug
		 FROM problem_tag_links ptl
		 INNER JOIN problem_tags pt ON pt.id = ptl.tag_id
		 WHERE ptl.problem_id = $1
		 ORDER BY pt.slug ASC`,
		problemID,
	)
	if err != nil {
		return nil, nil, nil, err
	}
	defer tagRows.Close()

	tags := make([]string, 0)
	for tagRows.Next() {
		var tag string
		if err := tagRows.Scan(&tag); err != nil {
			return nil, nil, nil, err
		}
		tags = append(tags, tag)
	}
	if err := tagRows.Err(); err != nil {
		return nil, nil, nil, err
	}

	return buildStatementSections(statementMD, inputMD, outputMD, constraintsMD), samples, tags, nil
}

func (r *SQLRepository) loadProblemPublicState(ctx context.Context, queries problemQueries, problemID uuid.UUID) (publicProblemState, error) {
	var state publicProblemState
	if err := queries.QueryRow(
		ctx,
		`SELECT
		   p.id,
		   p.problem_no,
		   p.slug,
		   p.title,
		   p.difficulty::text,
		   p.time_limit_ms,
		   p.memory_limit_kb,
		   p.is_published,
		   p.current_public_version_id,
		   p.current_judge_version_id,
		   COALESCE(ps.accepted_rate::float8, 0)
		 FROM problems p
		 LEFT JOIN problem_stats ps ON ps.problem_id = p.id
		 WHERE p.id = $1`,
		problemID,
	).Scan(
		&state.ID,
		&state.ProblemNo,
		&state.Slug,
		&state.Title,
		&state.Difficulty,
		&state.TimeLimitMs,
		&state.MemoryLimitKB,
		&state.IsPublished,
		&state.PublicVersionID,
		&state.JudgeVersionID,
		&state.AcceptedRate,
	); err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return publicProblemState{}, errors.New("problem not found")
		}
		return publicProblemState{}, err
	}

	statements, samples, tags, err := r.loadProblemContent(ctx, queries, problemID)
	if err != nil {
		return publicProblemState{}, err
	}
	state.StatementJSON = statements
	state.Samples = samples
	state.Tags = tags
	return state, nil
}

func (r *SQLRepository) writeSampleObjects(ctx context.Context, problemID uuid.UUID, samples []Sample) ([]storedSample, error) {
	stored := make([]storedSample, 0, len(samples))
	for index, sample := range samples {
		inputKey := fmt.Sprintf("problems/%s/samples/sample-%d.in", problemID, index+1)
		outputKey := fmt.Sprintf("problems/%s/samples/sample-%d.out", problemID, index+1)
		if err := r.store.PutObject(ctx, inputKey, []byte(sample.Input)); err != nil {
			return nil, err
		}
		if err := r.store.PutObject(ctx, outputKey, []byte(sample.Output)); err != nil {
			return nil, err
		}
		stored = append(stored, storedSample{
			InputObjectKey:  inputKey,
			OutputObjectKey: outputKey,
			Weight:          sample.Weight,
		})
	}
	return stored, nil
}

func buildStatementSections(statementMD string, inputMD string, outputMD string, constraintsMD string) []StatementSection {
	return []StatementSection{
		{Kind: "markdown", Section: "statement", Content: statementMD},
		{Kind: "markdown", Section: "input", Content: inputMD},
		{Kind: "markdown", Section: "output", Content: outputMD},
		{Kind: "markdown", Section: "constraints", Content: constraintsMD},
	}
}

func statementContent(items []StatementSection, section string) string {
	for _, item := range items {
		if item.Section == section {
			return item.Content
		}
	}
	return ""
}

func nullableUUIDStringValue(value *uuid.UUID) any {
	if value == nil {
		return nil
	}
	return value.String()
}
