-- Find contest public details that still contain legacy problem snapshots
-- without a slug field. These contests should be re-frozen.
SELECT
  c.id,
  c.slug,
  c.title,
  cpd.updated_at
FROM contest_public_details AS cpd
INNER JOIN contests AS c
  ON c.id = cpd.contest_id
WHERE EXISTS (
  SELECT 1
  FROM jsonb_array_elements(cpd.problems_json) AS problem
  WHERE COALESCE(NULLIF(problem ->> 'slug', ''), '') = ''
)
ORDER BY c.slug ASC;

-- Post-fix validation: should return 0 once all affected contests are re-frozen.
SELECT COUNT(*) AS contests_missing_problem_slug
FROM contest_public_details AS cpd
WHERE EXISTS (
  SELECT 1
  FROM jsonb_array_elements(cpd.problems_json) AS problem
  WHERE COALESCE(NULLIF(problem ->> 'slug', ''), '') = ''
);
