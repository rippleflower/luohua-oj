-- name: CreateSubmission :one
INSERT INTO submissions (id, user_id, problem_id, language, source_object, status)
VALUES ($1, $2, $3, $4, $5, 'PENDING')
RETURNING *;

-- name: GetSubmission :one
SELECT *
FROM submissions
WHERE id = $1;

-- name: UpdateSubmissionStatus :one
UPDATE submissions
SET
  status = $2,
  score = $3,
  compile_output = $4,
  max_time_ms = $5,
  max_memory_kb = $6,
  judged_at = now()
WHERE id = $1
RETURNING *;
