-- +goose Up
CREATE INDEX IF NOT EXISTS submissions_contest_user_problem_created_idx
  ON submissions (contest_id, user_id, problem_id, created_at DESC, id DESC);

-- +goose Down
DROP INDEX IF EXISTS submissions_contest_user_problem_created_idx;
