-- name: CreateProblem :one
INSERT INTO problems (
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
)
VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10)
RETURNING *;
