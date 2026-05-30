package contest

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/example/oj3/apps/api/internal/auth"
	db "github.com/example/oj3/apps/api/internal/db/generated"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/jackc/pgx/v5/pgxpool"
)

type SQLRepository struct {
	db *pgxpool.Pool
}

func NewSQLRepository(database db.DBTX) *SQLRepository {
	pool, ok := database.(*pgxpool.Pool)
	if !ok {
		panic("contest.NewSQLRepository requires *pgxpool.Pool")
	}
	return &SQLRepository{db: pool}
}

func (r *SQLRepository) ListContests(ctx context.Context) ([]Summary, error) {
	rows, err := r.db.Query(
		ctx,
		`SELECT contest_id, slug, title, status, starts_at, ends_at, duration_label, problem_count, participant_count, blurb
		 FROM contest_public_summaries
		 ORDER BY starts_at DESC, slug ASC`,
	)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var contests []Summary
	for rows.Next() {
		var item Summary
		if err := rows.Scan(
			&item.ID,
			&item.Slug,
			&item.Title,
			&item.Status,
			&item.StartsAt,
			&item.EndsAt,
			&item.DurationLabel,
			&item.ProblemCount,
			&item.ParticipantCount,
			&item.Blurb,
		); err != nil {
			return nil, err
		}
		contests = append(contests, item)
	}

	return contests, rows.Err()
}

func (r *SQLRepository) GetContestBySlug(ctx context.Context, slug string) (Detail, error) {
	var (
		item                  Detail
		recentSubmissionsJSON []byte
		problemsJSON          []byte
	)
	err := r.db.QueryRow(
		ctx,
		`SELECT
		   contest_id,
		   slug,
		   title,
		   status,
		   starts_at,
		   ends_at,
		   duration_label,
		   problem_count,
		   participant_count,
		   blurb,
		   rank_summary,
		   remaining_label,
		   recent_submissions_json,
		   problems_json,
		   updated_at
		 FROM contest_public_details
		 WHERE slug = $1`,
		slug,
	).Scan(
		&item.ID,
		&item.Slug,
		&item.Title,
		&item.Status,
		&item.StartsAt,
		&item.EndsAt,
		&item.DurationLabel,
		&item.ProblemCount,
		&item.ParticipantCount,
		&item.Blurb,
		&item.RankSummary,
		&item.Remaining,
		&recentSubmissionsJSON,
		&problemsJSON,
		&item.UpdatedAt,
	)
	if err != nil {
		return Detail{}, err
	}
	if err := json.Unmarshal(recentSubmissionsJSON, &item.RecentSubmissions); err != nil {
		return Detail{}, err
	}
	if err := json.Unmarshal(problemsJSON, &item.Problems); err != nil {
		return Detail{}, err
	}
	return item, nil
}

func (r *SQLRepository) ListAdminContests(ctx context.Context) ([]AdminContest, error) {
	rows, err := r.db.Query(
		ctx,
		`SELECT
		   c.id,
		   c.slug,
		   c.title,
		   c.blurb,
		   c.status,
		   c.starts_at,
		   c.ends_at,
		   c.participant_count,
		   COALESCE((SELECT COUNT(*) FROM contest_problem_links cpl WHERE cpl.contest_id = c.id), 0),
		   COALESCE((SELECT MAX(snapshot_no) FROM contest_snapshots cs WHERE cs.contest_id = c.id), 0),
		   c.updated_at
		 FROM contests c
		 ORDER BY c.starts_at DESC, c.slug ASC`,
	)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	items := make([]AdminContest, 0)
	contestIDs := make([]uuid.UUID, 0)
	byID := make(map[uuid.UUID]int)
	for rows.Next() {
		var item AdminContest
		if err := rows.Scan(
			&item.ID,
			&item.Slug,
			&item.Title,
			&item.Description,
			&item.Status,
			&item.StartsAt,
			&item.EndsAt,
			&item.ParticipantCount,
			&item.ProblemCount,
			&item.LatestSnapshotNo,
			&item.UpdatedAt,
		); err != nil {
			return nil, err
		}
		byID[item.ID] = len(items)
		contestIDs = append(contestIDs, item.ID)
		items = append(items, item)
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}
	if len(contestIDs) == 0 {
		return items, nil
	}

	if err := r.loadAdminContestProblems(ctx, contestIDs, items, byID); err != nil {
		return nil, err
	}
	if err := r.loadAdminContestSnapshots(ctx, contestIDs, items, byID); err != nil {
		return nil, err
	}

	return items, nil
}

func (r *SQLRepository) CreateAdminContest(ctx context.Context, actor auth.AuthenticatedUser, input CreateAdminInput) (AdminContest, error) {
	tx, err := r.db.Begin(ctx)
	if err != nil {
		return AdminContest{}, err
	}
	defer tx.Rollback(ctx)

	var contestID uuid.UUID
	err = tx.QueryRow(
		ctx,
		`INSERT INTO contests (slug, title, status, starts_at, ends_at, blurb, rank_summary, remaining_label)
		 VALUES ($1, $2, $3, $4, $5, $6, '', '')
		 RETURNING id`,
		input.Slug,
		input.Title,
		input.Status,
		input.StartsAt,
		input.EndsAt,
		input.Description,
	).Scan(&contestID)
	if err != nil {
		return AdminContest{}, mapContestWriteError(err)
	}

	if err := recordContestAudit(ctx, tx, actor, "admin.contest.created", contestID.String(), map[string]any{
		"slug":        input.Slug,
		"title":       input.Title,
		"description": input.Description,
		"status":      input.Status,
		"startsAt":    input.StartsAt.UTC().Format(time.RFC3339),
		"endsAt":      input.EndsAt.UTC().Format(time.RFC3339),
	}, input.Reason, input.IP); err != nil {
		return AdminContest{}, err
	}

	if err := tx.Commit(ctx); err != nil {
		return AdminContest{}, err
	}
	return r.getAdminContest(ctx, contestID)
}

func (r *SQLRepository) UpdateAdminContest(ctx context.Context, actor auth.AuthenticatedUser, input UpdateAdminInput) (AdminContest, error) {
	tx, err := r.db.Begin(ctx)
	if err != nil {
		return AdminContest{}, err
	}
	defer tx.Rollback(ctx)

	tag, err := tx.Exec(
		ctx,
		`UPDATE contests
		 SET slug = $2,
		     title = $3,
		     status = $4,
		     starts_at = $5,
		     ends_at = $6,
		     blurb = $7,
		     updated_at = now()
		 WHERE id = $1`,
		input.ContestID,
		input.Slug,
		input.Title,
		input.Status,
		input.StartsAt,
		input.EndsAt,
		input.Description,
	)
	if err != nil {
		return AdminContest{}, mapContestWriteError(err)
	}
	if tag.RowsAffected() == 0 {
		return AdminContest{}, errors.New("contest not found")
	}

	if err := recordContestAudit(ctx, tx, actor, "admin.contest.updated", input.ContestID.String(), map[string]any{
		"slug":        input.Slug,
		"title":       input.Title,
		"description": input.Description,
		"status":      input.Status,
		"startsAt":    input.StartsAt.UTC().Format(time.RFC3339),
		"endsAt":      input.EndsAt.UTC().Format(time.RFC3339),
	}, input.Reason, input.IP); err != nil {
		return AdminContest{}, err
	}

	if err := tx.Commit(ctx); err != nil {
		return AdminContest{}, err
	}
	return r.getAdminContest(ctx, input.ContestID)
}

func (r *SQLRepository) ReplaceAdminContestProblems(ctx context.Context, actor auth.AuthenticatedUser, input ReplaceProblemsInput) (AdminContest, error) {
	tx, err := r.db.Begin(ctx)
	if err != nil {
		return AdminContest{}, err
	}
	defer tx.Rollback(ctx)

	if err := ensureContestExists(ctx, tx, input.ContestID); err != nil {
		return AdminContest{}, err
	}
	if err := ensureProblemsExist(ctx, tx, input.Problems); err != nil {
		return AdminContest{}, err
	}

	if _, err := tx.Exec(ctx, `DELETE FROM contest_problem_links WHERE contest_id = $1`, input.ContestID); err != nil {
		return AdminContest{}, err
	}
	for _, item := range input.Problems {
		if _, err := tx.Exec(
			ctx,
			`INSERT INTO contest_problem_links (contest_id, problem_id, code, position)
			 VALUES ($1, $2, $3, $4)`,
			input.ContestID,
			item.ProblemID,
			item.Code,
			item.Position,
		); err != nil {
			return AdminContest{}, mapContestWriteError(err)
		}
	}
	if _, err := tx.Exec(ctx, `UPDATE contests SET updated_at = now() WHERE id = $1`, input.ContestID); err != nil {
		return AdminContest{}, err
	}

	diffProblems := make([]map[string]any, 0, len(input.Problems))
	for _, item := range input.Problems {
		diffProblems = append(diffProblems, map[string]any{
			"problemId": item.ProblemID.String(),
			"code":      item.Code,
			"position":  item.Position,
		})
	}
	if err := recordContestAudit(ctx, tx, actor, "admin.contest.problems.replaced", input.ContestID.String(), map[string]any{
		"problems": diffProblems,
	}, input.Reason, input.IP); err != nil {
		return AdminContest{}, err
	}

	if err := tx.Commit(ctx); err != nil {
		return AdminContest{}, err
	}
	return r.getAdminContest(ctx, input.ContestID)
}

func (r *SQLRepository) FreezeAdminContest(ctx context.Context, actor auth.AuthenticatedUser, input FreezeAdminInput) (AdminContest, error) {
	tx, err := r.db.Begin(ctx)
	if err != nil {
		return AdminContest{}, err
	}
	defer tx.Rollback(ctx)

	base, problems, snapshotNo, err := r.loadContestFreezeState(ctx, tx, input.ContestID)
	if err != nil {
		return AdminContest{}, err
	}
	if len(problems) == 0 {
		return AdminContest{}, errors.New("contest has no problems")
	}

	detailJSON, err := json.Marshal(map[string]any{
		"slug":        base.Slug,
		"title":       base.Title,
		"description": base.Description,
		"status":      base.Status,
		"startsAt":    base.StartsAt.UTC().Format(time.RFC3339),
		"endsAt":      base.EndsAt.UTC().Format(time.RFC3339),
	})
	if err != nil {
		return AdminContest{}, err
	}

	snapshotID := uuid.New()
	objectKey := fmt.Sprintf("contests/%s/snapshots/%d.json", base.ID, snapshotNo)
	if _, err := tx.Exec(
		ctx,
		`INSERT INTO contest_snapshots (
		   id,
		   contest_id,
		   snapshot_no,
		   detail_json,
		   scoreboard_config_json,
		   object_key
		 ) VALUES ($1, $2, $3, $4, '{}'::jsonb, $5)`,
		snapshotID,
		base.ID,
		snapshotNo,
		detailJSON,
		objectKey,
	); err != nil {
		return AdminContest{}, err
	}

	for _, item := range problems {
		limitsJSON, err := json.Marshal(map[string]any{
			"timeLimitMs":   item.TimeLimitMs,
			"memoryLimitKb": item.MemoryLimitKb,
		})
		if err != nil {
			return AdminContest{}, err
		}
		metadataJSON, err := json.Marshal(map[string]any{
			"position":         item.Position,
			"problemSlug":      item.ProblemSlug,
			"problemVersionId": nullableUUIDString(item.ProblemVersionID),
		})
		if err != nil {
			return AdminContest{}, err
		}
		if _, err := tx.Exec(
			ctx,
			`INSERT INTO contest_problem_snapshots (
			   contest_id,
			   snapshot_id,
			   problem_id,
			   problem_version_id,
			   code,
			   title,
			   difficulty,
			   statement_excerpt,
			   limits_json,
			   metadata_json
			 ) VALUES ($1, $2, $3, $4, $5, $6, $7::problem_difficulty, $8, $9, $10)`,
			base.ID,
			snapshotID,
			item.ProblemID,
			item.ProblemVersionID,
			item.Code,
			item.ProblemTitle,
			item.Difficulty,
			trimExcerpt(item.Statement),
			limitsJSON,
			metadataJSON,
		); err != nil {
			return AdminContest{}, err
		}
	}

	recentSubmissions, err := loadContestRecentSubmissions(ctx, tx, base.ID)
	if err != nil {
		return AdminContest{}, err
	}
	recentSubmissionsJSON, err := json.Marshal(recentSubmissions)
	if err != nil {
		return AdminContest{}, err
	}
	publicProblems := make([]ProblemSnapshot, 0, len(problems))
	for _, item := range problems {
		publicProblems = append(publicProblems, ProblemSnapshot{
			Code:       item.Code,
			Title:      item.ProblemTitle,
			Difficulty: item.Difficulty,
			Status:     "LOCKED",
		})
	}
	problemsJSON, err := json.Marshal(publicProblems)
	if err != nil {
		return AdminContest{}, err
	}

	durationLabel := formatContestDuration(base.StartsAt, base.EndsAt)
	remainingLabel := deriveRemainingLabel(base.Status, base.StartsAt, base.EndsAt, time.Now().UTC())
	if _, err := tx.Exec(
		ctx,
		`INSERT INTO contest_public_summaries (
		   contest_id,
		   slug,
		   title,
		   status,
		   starts_at,
		   ends_at,
		   duration_label,
		   problem_count,
		   participant_count,
		   blurb,
		   updated_at
		 ) VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, now())
		 ON CONFLICT (contest_id) DO UPDATE SET
		   slug = EXCLUDED.slug,
		   title = EXCLUDED.title,
		   status = EXCLUDED.status,
		   starts_at = EXCLUDED.starts_at,
		   ends_at = EXCLUDED.ends_at,
		   duration_label = EXCLUDED.duration_label,
		   problem_count = EXCLUDED.problem_count,
		   participant_count = EXCLUDED.participant_count,
		   blurb = EXCLUDED.blurb,
		   updated_at = EXCLUDED.updated_at`,
		base.ID,
		base.Slug,
		base.Title,
		base.Status,
		base.StartsAt,
		base.EndsAt,
		durationLabel,
		len(problems),
		base.ParticipantCount,
		base.Description,
	); err != nil {
		return AdminContest{}, err
	}
	if _, err := tx.Exec(
		ctx,
		`INSERT INTO contest_public_details (
		   contest_id,
		   slug,
		   title,
		   status,
		   starts_at,
		   ends_at,
		   duration_label,
		   problem_count,
		   participant_count,
		   blurb,
		   rank_summary,
		   remaining_label,
		   recent_submissions_json,
		   problems_json,
		   updated_at
		 ) VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13, $14, now())
		 ON CONFLICT (contest_id) DO UPDATE SET
		   slug = EXCLUDED.slug,
		   title = EXCLUDED.title,
		   status = EXCLUDED.status,
		   starts_at = EXCLUDED.starts_at,
		   ends_at = EXCLUDED.ends_at,
		   duration_label = EXCLUDED.duration_label,
		   problem_count = EXCLUDED.problem_count,
		   participant_count = EXCLUDED.participant_count,
		   blurb = EXCLUDED.blurb,
		   rank_summary = EXCLUDED.rank_summary,
		   remaining_label = EXCLUDED.remaining_label,
		   recent_submissions_json = EXCLUDED.recent_submissions_json,
		   problems_json = EXCLUDED.problems_json,
		   updated_at = EXCLUDED.updated_at`,
		base.ID,
		base.Slug,
		base.Title,
		base.Status,
		base.StartsAt,
		base.EndsAt,
		durationLabel,
		len(problems),
		base.ParticipantCount,
		base.Description,
		base.RankSummary,
		remainingLabel,
		recentSubmissionsJSON,
		problemsJSON,
	); err != nil {
		return AdminContest{}, err
	}

	if err := recordContestAudit(ctx, tx, actor, "admin.contest.frozen", base.ID.String(), map[string]any{
		"snapshotNo":   snapshotNo,
		"snapshotId":   snapshotID.String(),
		"problemCount": len(problems),
	}, input.Reason, input.IP); err != nil {
		return AdminContest{}, err
	}
	if _, err := tx.Exec(ctx, `UPDATE contests SET updated_at = now() WHERE id = $1`, base.ID); err != nil {
		return AdminContest{}, err
	}

	if err := tx.Commit(ctx); err != nil {
		return AdminContest{}, err
	}
	return r.getAdminContest(ctx, base.ID)
}

type contestBase struct {
	ID               uuid.UUID
	Slug             string
	Title            string
	Description      string
	Status           string
	StartsAt         time.Time
	EndsAt           time.Time
	ParticipantCount int
	UpdatedAt        time.Time
	RankSummary      string
}

type contestFreezeProblem struct {
	ProblemID        uuid.UUID
	ProblemVersionID *uuid.UUID
	ProblemSlug      string
	ProblemTitle     string
	Code             string
	Position         int
	Difficulty       string
	Statement        string
	TimeLimitMs      int
	MemoryLimitKb    int
}

func (r *SQLRepository) getAdminContest(ctx context.Context, contestID uuid.UUID) (AdminContest, error) {
	items, err := r.loadAdminContestsByID(ctx, []uuid.UUID{contestID})
	if err != nil {
		return AdminContest{}, err
	}
	if len(items) == 0 {
		return AdminContest{}, errors.New("contest not found")
	}
	return items[0], nil
}

func (r *SQLRepository) loadAdminContestsByID(ctx context.Context, contestIDs []uuid.UUID) ([]AdminContest, error) {
	rows, err := r.db.Query(
		ctx,
		`SELECT
		   c.id,
		   c.slug,
		   c.title,
		   c.blurb,
		   c.status,
		   c.starts_at,
		   c.ends_at,
		   c.participant_count,
		   COALESCE((SELECT COUNT(*) FROM contest_problem_links cpl WHERE cpl.contest_id = c.id), 0),
		   COALESCE((SELECT MAX(snapshot_no) FROM contest_snapshots cs WHERE cs.contest_id = c.id), 0),
		   c.updated_at
		 FROM contests c
		 WHERE c.id = ANY($1::uuid[])
		 ORDER BY c.starts_at DESC, c.slug ASC`,
		contestIDs,
	)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	items := make([]AdminContest, 0)
	byID := make(map[uuid.UUID]int)
	for rows.Next() {
		var item AdminContest
		if err := rows.Scan(
			&item.ID,
			&item.Slug,
			&item.Title,
			&item.Description,
			&item.Status,
			&item.StartsAt,
			&item.EndsAt,
			&item.ParticipantCount,
			&item.ProblemCount,
			&item.LatestSnapshotNo,
			&item.UpdatedAt,
		); err != nil {
			return nil, err
		}
		byID[item.ID] = len(items)
		items = append(items, item)
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}
	if len(items) == 0 {
		return items, nil
	}

	if err := r.loadAdminContestProblems(ctx, contestIDs, items, byID); err != nil {
		return nil, err
	}
	if err := r.loadAdminContestSnapshots(ctx, contestIDs, items, byID); err != nil {
		return nil, err
	}
	return items, nil
}

func (r *SQLRepository) loadAdminContestProblems(ctx context.Context, contestIDs []uuid.UUID, items []AdminContest, byID map[uuid.UUID]int) error {
	rows, err := r.db.Query(
		ctx,
		`SELECT
		   cpl.contest_id,
		   p.id,
		   p.slug,
		   p.title,
		   cpl.code,
		   cpl.position
		 FROM contest_problem_links cpl
		 INNER JOIN problems p ON p.id = cpl.problem_id
		 WHERE cpl.contest_id = ANY($1::uuid[])
		 ORDER BY cpl.contest_id ASC, cpl.position ASC, cpl.code ASC`,
		contestIDs,
	)
	if err != nil {
		return err
	}
	defer rows.Close()

	for rows.Next() {
		var (
			contestID uuid.UUID
			item      AdminProblemBinding
		)
		if err := rows.Scan(&contestID, &item.ProblemID, &item.ProblemSlug, &item.ProblemTitle, &item.Code, &item.Position); err != nil {
			return err
		}
		index, ok := byID[contestID]
		if !ok {
			continue
		}
		items[index].Problems = append(items[index].Problems, item)
	}
	return rows.Err()
}

func (r *SQLRepository) loadAdminContestSnapshots(ctx context.Context, contestIDs []uuid.UUID, items []AdminContest, byID map[uuid.UUID]int) error {
	rows, err := r.db.Query(
		ctx,
		`SELECT
		   cs.contest_id,
		   cs.id,
		   cs.snapshot_no,
		   cs.frozen_at,
		   COALESCE(COUNT(cps.*), 0)::int
		 FROM contest_snapshots cs
		 LEFT JOIN contest_problem_snapshots cps ON cps.snapshot_id = cs.id
		 WHERE cs.contest_id = ANY($1::uuid[])
		 GROUP BY cs.contest_id, cs.id, cs.snapshot_no, cs.frozen_at
		 ORDER BY cs.contest_id ASC, cs.snapshot_no DESC`,
		contestIDs,
	)
	if err != nil {
		return err
	}
	defer rows.Close()

	for rows.Next() {
		var contestID uuid.UUID
		var snapshot AdminSnapshotSummary
		if err := rows.Scan(&contestID, &snapshot.ID, &snapshot.SnapshotNo, &snapshot.FrozenAt, &snapshot.ProblemCount); err != nil {
			return err
		}
		index, ok := byID[contestID]
		if !ok {
			continue
		}
		items[index].Snapshots = append(items[index].Snapshots, snapshot)
	}
	return rows.Err()
}

func (r *SQLRepository) loadContestFreezeState(ctx context.Context, tx pgx.Tx, contestID uuid.UUID) (contestBase, []contestFreezeProblem, int, error) {
	var base contestBase
	err := tx.QueryRow(
		ctx,
		`SELECT id, slug, title, blurb, status, starts_at, ends_at, participant_count, updated_at, rank_summary
		 FROM contests
		 WHERE id = $1
		 FOR UPDATE`,
		contestID,
	).Scan(
		&base.ID,
		&base.Slug,
		&base.Title,
		&base.Description,
		&base.Status,
		&base.StartsAt,
		&base.EndsAt,
		&base.ParticipantCount,
		&base.UpdatedAt,
		&base.RankSummary,
	)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return contestBase{}, nil, 0, errors.New("contest not found")
		}
		return contestBase{}, nil, 0, err
	}

	rows, err := tx.Query(
		ctx,
		`SELECT
		   p.id,
		   p.current_judge_version_id,
		   p.slug,
		   p.title,
		   cpl.code,
		   cpl.position,
		   p.difficulty::text,
		   p.statement_md,
		   p.time_limit_ms,
		   p.memory_limit_kb
		 FROM contest_problem_links cpl
		 INNER JOIN problems p ON p.id = cpl.problem_id
		 WHERE cpl.contest_id = $1
		 ORDER BY cpl.position ASC, cpl.code ASC`,
		contestID,
	)
	if err != nil {
		return contestBase{}, nil, 0, err
	}
	defer rows.Close()

	problems := make([]contestFreezeProblem, 0)
	for rows.Next() {
		var (
			item      contestFreezeProblem
			versionID uuid.UUID
		)
		if err := rows.Scan(
			&item.ProblemID,
			&versionID,
			&item.ProblemSlug,
			&item.ProblemTitle,
			&item.Code,
			&item.Position,
			&item.Difficulty,
			&item.Statement,
			&item.TimeLimitMs,
			&item.MemoryLimitKb,
		); err != nil {
			return contestBase{}, nil, 0, err
		}
		if versionID != uuid.Nil {
			item.ProblemVersionID = &versionID
		}
		problems = append(problems, item)
	}
	if err := rows.Err(); err != nil {
		return contestBase{}, nil, 0, err
	}

	var currentSnapshotNo int
	if err := tx.QueryRow(
		ctx,
		`SELECT COALESCE(MAX(snapshot_no), 0)
		 FROM contest_snapshots
		 WHERE contest_id = $1`,
		contestID,
	).Scan(&currentSnapshotNo); err != nil {
		return contestBase{}, nil, 0, err
	}

	return base, problems, currentSnapshotNo + 1, nil
}

func ensureContestExists(ctx context.Context, tx pgx.Tx, contestID uuid.UUID) error {
	var found uuid.UUID
	if err := tx.QueryRow(ctx, `SELECT id FROM contests WHERE id = $1`, contestID).Scan(&found); err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return errors.New("contest not found")
		}
		return err
	}
	return nil
}

func ensureProblemsExist(ctx context.Context, tx pgx.Tx, problems []AdminProblemBindingInput) error {
	if len(problems) == 0 {
		return nil
	}
	ids := make([]uuid.UUID, 0, len(problems))
	for _, item := range problems {
		ids = append(ids, item.ProblemID)
	}
	var count int
	if err := tx.QueryRow(ctx, `SELECT COUNT(*) FROM problems WHERE id = ANY($1::uuid[])`, ids).Scan(&count); err != nil {
		return err
	}
	if count != len(problems) {
		return errors.New("problem not found")
	}
	return nil
}

func loadContestRecentSubmissions(ctx context.Context, tx pgx.Tx, contestID uuid.UUID) ([]RecentSubmission, error) {
	rows, err := tx.Query(
		ctx,
		`SELECT id, status::text, created_at
		 FROM submissions
		 WHERE contest_id = $1
		 ORDER BY created_at DESC
		 LIMIT 8`,
		contestID,
	)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	items := make([]RecentSubmission, 0)
	for rows.Next() {
		var (
			item      RecentSubmission
			id        uuid.UUID
			createdAt time.Time
		)
		if err := rows.Scan(&id, &item.Status, &createdAt); err != nil {
			return nil, err
		}
		item.ID = id.String()
		item.ProblemCode = ""
		item.At = createdAt.UTC().Format(time.RFC3339)
		items = append(items, item)
	}
	return items, rows.Err()
}

func mapContestWriteError(err error) error {
	var pgErr *pgconn.PgError
	if !errors.As(err, &pgErr) {
		return err
	}
	switch pgErr.Code {
	case "23505":
		switch pgErr.ConstraintName {
		case "contests_slug_key":
			return errors.New("slug already exists")
		case "contest_problem_links_pkey":
			return errors.New("duplicate code")
		case "contest_problem_links_contest_id_position_key":
			return errors.New("duplicate position")
		}
	case "23503":
		if strings.Contains(pgErr.ConstraintName, "contest_problem_links_problem_id") {
			return errors.New("problem not found")
		}
	case "23514":
		if strings.Contains(pgErr.ConstraintName, "position") {
			return errors.New("position must be positive")
		}
	}
	return err
}

func recordContestAudit(ctx context.Context, tx pgx.Tx, actor auth.AuthenticatedUser, action string, targetID string, diff map[string]any, reason string, ip string) error {
	diffJSON, err := json.Marshal(diff)
	if err != nil {
		return err
	}
	_, err = tx.Exec(
		ctx,
		`INSERT INTO audit_logs (actor_user_id, actor_role, action, target_type, target_id, diff_json, reason, ip)
		 VALUES ($1, $2, $3, 'contest', $4, $5, $6, $7)`,
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

func trimExcerpt(value string) string {
	value = strings.TrimSpace(value)
	if len(value) <= 240 {
		return value
	}
	return strings.TrimSpace(value[:240])
}

func formatContestDuration(startsAt time.Time, endsAt time.Time) string {
	totalMinutes := int(endsAt.Sub(startsAt).Minutes())
	if totalMinutes <= 0 {
		return "0 minutes"
	}
	hours := totalMinutes / 60
	minutes := totalMinutes % 60
	switch {
	case hours > 0 && minutes == 0:
		if hours == 1 {
			return "1 hour"
		}
		return fmt.Sprintf("%d hours", hours)
	case hours == 0:
		if minutes == 1 {
			return "1 minute"
		}
		return fmt.Sprintf("%d minutes", minutes)
	default:
		return fmt.Sprintf("%dh %dm", hours, minutes)
	}
}

func deriveRemainingLabel(status string, startsAt time.Time, endsAt time.Time, now time.Time) string {
	switch status {
	case "UPCOMING":
		if now.Before(startsAt) {
			return fmt.Sprintf("starts in %s", humanDuration(startsAt.Sub(now)))
		}
		return "awaiting start"
	case "RUNNING":
		if now.Before(endsAt) {
			return formatClockDuration(endsAt.Sub(now))
		}
		return "contest ended"
	default:
		return "contest ended"
	}
}

func humanDuration(duration time.Duration) string {
	if duration < 0 {
		duration = 0
	}
	totalMinutes := int(duration.Round(time.Minute).Minutes())
	if totalMinutes < 60 {
		return fmt.Sprintf("%d minutes", totalMinutes)
	}
	hours := totalMinutes / 60
	minutes := totalMinutes % 60
	if minutes == 0 {
		return fmt.Sprintf("%d hours", hours)
	}
	return fmt.Sprintf("%dh %dm", hours, minutes)
}

func formatClockDuration(duration time.Duration) string {
	if duration < 0 {
		duration = 0
	}
	totalSeconds := int(duration.Round(time.Second).Seconds())
	hours := totalSeconds / 3600
	minutes := (totalSeconds % 3600) / 60
	seconds := totalSeconds % 60
	return fmt.Sprintf("%02d:%02d:%02d", hours, minutes, seconds)
}

func nullableUUIDString(value *uuid.UUID) string {
	if value == nil {
		return ""
	}
	return value.String()
}
