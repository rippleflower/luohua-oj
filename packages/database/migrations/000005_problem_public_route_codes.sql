-- +goose Up
CREATE SEQUENCE IF NOT EXISTS problem_no_seq;

ALTER TABLE problems
  ADD COLUMN IF NOT EXISTS problem_no BIGINT;

ALTER TABLE problems
  ALTER COLUMN problem_no SET DEFAULT nextval('problem_no_seq');

UPDATE problems
SET problem_no = nextval('problem_no_seq')
WHERE problem_no IS NULL;

SELECT setval(
  'problem_no_seq',
  GREATEST(COALESCE((SELECT MAX(problem_no) FROM problems), 0), 1),
  true
);

ALTER TABLE problems
  ALTER COLUMN problem_no SET NOT NULL;

ALTER TABLE problems
  ADD CONSTRAINT problems_problem_no_key UNIQUE (problem_no);

ALTER TABLE problem_public_summaries
  ADD COLUMN IF NOT EXISTS problem_no BIGINT;

ALTER TABLE problem_public_details
  ADD COLUMN IF NOT EXISTS problem_no BIGINT;

UPDATE problem_public_summaries pps
SET problem_no = p.problem_no
FROM problems p
WHERE p.id = pps.problem_id AND pps.problem_no IS NULL;

UPDATE problem_public_details ppd
SET problem_no = p.problem_no
FROM problems p
WHERE p.id = ppd.problem_id AND ppd.problem_no IS NULL;

ALTER TABLE problem_public_summaries
  ALTER COLUMN problem_no SET NOT NULL;

ALTER TABLE problem_public_details
  ALTER COLUMN problem_no SET NOT NULL;

CREATE UNIQUE INDEX IF NOT EXISTS problem_public_summaries_problem_no_idx
  ON problem_public_summaries (problem_no);

CREATE UNIQUE INDEX IF NOT EXISTS problem_public_details_problem_no_idx
  ON problem_public_details (problem_no);

-- +goose Down
DROP INDEX IF EXISTS problem_public_details_problem_no_idx;
DROP INDEX IF EXISTS problem_public_summaries_problem_no_idx;

ALTER TABLE problem_public_details
  DROP COLUMN IF EXISTS problem_no;

ALTER TABLE problem_public_summaries
  DROP COLUMN IF EXISTS problem_no;

ALTER TABLE problems
  DROP CONSTRAINT IF EXISTS problems_problem_no_key;

ALTER TABLE problems
  DROP COLUMN IF EXISTS problem_no;

DROP SEQUENCE IF EXISTS problem_no_seq;
