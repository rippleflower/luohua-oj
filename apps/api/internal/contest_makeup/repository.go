package contest_makeup

import (
	"context"
	"errors"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

type SQLRepository struct {
	db *pgxpool.Pool
}

func NewSQLRepository(pool *pgxpool.Pool) *SQLRepository {
	return &SQLRepository{db: pool}
}

func (r *SQLRepository) GetContestMeta(ctx context.Context, slug string) (ContestMeta, error) {
	var meta ContestMeta
	err := r.db.QueryRow(
		ctx,
		`SELECT id, slug, status
		 FROM contests
		 WHERE slug = $1`,
		slug,
	).Scan(&meta.ID, &meta.Slug, &meta.Status)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return ContestMeta{}, ErrContestNotFound
		}
		return ContestMeta{}, err
	}
	return meta, nil
}

func (r *SQLRepository) ListContestProblems(ctx context.Context, contestID uuid.UUID) ([]ContestProblem, error) {
	rows, err := r.db.Query(
		ctx,
		`SELECT cpl.problem_id, cpl.code, cpl.position, p.slug, p.title, p.difficulty::text
		 FROM contest_problem_links cpl
		 INNER JOIN problems p ON p.id = cpl.problem_id
		 WHERE cpl.contest_id = $1
		 ORDER BY cpl.position ASC, cpl.code ASC`,
		contestID,
	)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	items := make([]ContestProblem, 0)
	for rows.Next() {
		var item ContestProblem
		if err := rows.Scan(
			&item.ProblemID,
			&item.ProblemCode,
			&item.Position,
			&item.ProblemSlug,
			&item.ProblemTitle,
			&item.Difficulty,
		); err != nil {
			return nil, err
		}
		items = append(items, item)
	}
	return items, rows.Err()
}

func (r *SQLRepository) ListUserProblemAttempts(ctx context.Context, userID uuid.UUID, contestID uuid.UUID) (map[uuid.UUID]ProblemAttempt, error) {
	rows, err := r.db.Query(
		ctx,
		`WITH filtered AS (
		   SELECT
		     s.problem_id,
		     s.status,
		     s.created_at,
		     s.id
		   FROM submissions s
		   WHERE s.user_id = $1
		     AND s.contest_id = $2
		 ),
		 aggregated AS (
		   SELECT
		     f.problem_id,
		     BOOL_OR(f.status = 'ACCEPTED') AS has_accepted,
		     COUNT(*)::int AS attempt_count
		   FROM filtered f
		   GROUP BY f.problem_id
		 ),
		 latest AS (
		   SELECT DISTINCT ON (f.problem_id)
		     f.problem_id,
		     f.status::text AS latest_status
		   FROM filtered f
		   ORDER BY f.problem_id, f.created_at DESC, f.id DESC
		 )
		 SELECT l.problem_id, l.latest_status, a.has_accepted, a.attempt_count
		 FROM latest l
		 INNER JOIN aggregated a ON a.problem_id = l.problem_id`,
		userID,
		contestID,
	)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	items := make(map[uuid.UUID]ProblemAttempt)
	for rows.Next() {
		var item ProblemAttempt
		if err := rows.Scan(&item.ProblemID, &item.LatestStatus, &item.HasAccepted, &item.AttemptCount); err != nil {
			return nil, err
		}
		items[item.ProblemID] = item
	}
	return items, rows.Err()
}
