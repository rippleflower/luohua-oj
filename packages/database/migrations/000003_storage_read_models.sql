-- +goose Up
ALTER TABLE users
  ADD COLUMN IF NOT EXISTS status TEXT NOT NULL DEFAULT 'ACTIVE';

CREATE TABLE IF NOT EXISTS problem_versions (
  id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
  problem_id UUID NOT NULL REFERENCES problems(id) ON DELETE CASCADE,
  version_no INTEGER NOT NULL CHECK (version_no > 0),
  manifest_object_key TEXT NOT NULL,
  public_snapshot_version INTEGER NOT NULL DEFAULT 1 CHECK (public_snapshot_version > 0),
  judge_bundle_version INTEGER NOT NULL DEFAULT 1 CHECK (judge_bundle_version > 0),
  status TEXT NOT NULL DEFAULT 'DRAFT',
  created_at TIMESTAMPTZ NOT NULL DEFAULT now(),
  updated_at TIMESTAMPTZ NOT NULL DEFAULT now(),
  UNIQUE (problem_id, version_no)
);

ALTER TABLE problems
  ADD COLUMN IF NOT EXISTS current_public_version_id UUID,
  ADD COLUMN IF NOT EXISTS current_judge_version_id UUID;

CREATE TABLE IF NOT EXISTS problem_tags (
  id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
  slug TEXT NOT NULL UNIQUE,
  label TEXT NOT NULL,
  created_at TIMESTAMPTZ NOT NULL DEFAULT now()
);

CREATE TABLE IF NOT EXISTS problem_tag_links (
  problem_id UUID NOT NULL REFERENCES problems(id) ON DELETE CASCADE,
  tag_id UUID NOT NULL REFERENCES problem_tags(id) ON DELETE CASCADE,
  PRIMARY KEY (problem_id, tag_id)
);

CREATE TABLE IF NOT EXISTS problem_stats (
  problem_id UUID PRIMARY KEY REFERENCES problems(id) ON DELETE CASCADE,
  submission_count INTEGER NOT NULL DEFAULT 0 CHECK (submission_count >= 0),
  accepted_count INTEGER NOT NULL DEFAULT 0 CHECK (accepted_count >= 0),
  accepted_rate NUMERIC(5,2) NOT NULL DEFAULT 0 CHECK (accepted_rate >= 0 AND accepted_rate <= 100),
  updated_at TIMESTAMPTZ NOT NULL DEFAULT now()
);

CREATE TABLE IF NOT EXISTS problem_public_summaries (
  problem_id UUID PRIMARY KEY REFERENCES problems(id) ON DELETE CASCADE,
  slug TEXT NOT NULL UNIQUE,
  title TEXT NOT NULL,
  difficulty problem_difficulty NOT NULL,
  tags_json JSONB NOT NULL DEFAULT '[]'::jsonb,
  accepted_rate NUMERIC(5,2) NOT NULL DEFAULT 0,
  statistics_json JSONB NOT NULL DEFAULT '{}'::jsonb,
  updated_at TIMESTAMPTZ NOT NULL DEFAULT now()
);

CREATE TABLE IF NOT EXISTS problem_public_details (
  problem_id UUID PRIMARY KEY REFERENCES problems(id) ON DELETE CASCADE,
  slug TEXT NOT NULL UNIQUE,
  title TEXT NOT NULL,
  difficulty problem_difficulty NOT NULL,
  statement_json JSONB NOT NULL,
  samples_json JSONB NOT NULL DEFAULT '[]'::jsonb,
  limits_json JSONB NOT NULL DEFAULT '{}'::jsonb,
  metadata_json JSONB NOT NULL DEFAULT '{}'::jsonb,
  updated_at TIMESTAMPTZ NOT NULL DEFAULT now()
);

CREATE TABLE IF NOT EXISTS contests (
  id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
  slug TEXT NOT NULL UNIQUE,
  title TEXT NOT NULL,
  status TEXT NOT NULL,
  starts_at TIMESTAMPTZ NOT NULL,
  ends_at TIMESTAMPTZ NOT NULL,
  blurb TEXT NOT NULL DEFAULT '',
  rank_summary TEXT NOT NULL DEFAULT '',
  remaining_label TEXT NOT NULL DEFAULT '',
  participant_count INTEGER NOT NULL DEFAULT 0 CHECK (participant_count >= 0),
  freeze_scoreboard BOOLEAN NOT NULL DEFAULT false,
  created_at TIMESTAMPTZ NOT NULL DEFAULT now(),
  updated_at TIMESTAMPTZ NOT NULL DEFAULT now()
);

CREATE TABLE IF NOT EXISTS contest_problem_links (
  contest_id UUID NOT NULL REFERENCES contests(id) ON DELETE CASCADE,
  problem_id UUID NOT NULL REFERENCES problems(id) ON DELETE CASCADE,
  code TEXT NOT NULL,
  position INTEGER NOT NULL CHECK (position > 0),
  PRIMARY KEY (contest_id, code),
  UNIQUE (contest_id, position)
);

CREATE TABLE IF NOT EXISTS contest_snapshots (
  id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
  contest_id UUID NOT NULL REFERENCES contests(id) ON DELETE CASCADE,
  snapshot_no INTEGER NOT NULL CHECK (snapshot_no > 0),
  frozen_at TIMESTAMPTZ NOT NULL DEFAULT now(),
  detail_json JSONB NOT NULL DEFAULT '{}'::jsonb,
  scoreboard_config_json JSONB NOT NULL DEFAULT '{}'::jsonb,
  object_key TEXT NOT NULL,
  created_at TIMESTAMPTZ NOT NULL DEFAULT now(),
  UNIQUE (contest_id, snapshot_no)
);

CREATE TABLE IF NOT EXISTS contest_problem_snapshots (
  contest_id UUID NOT NULL REFERENCES contests(id) ON DELETE CASCADE,
  snapshot_id UUID NOT NULL REFERENCES contest_snapshots(id) ON DELETE CASCADE,
  problem_id UUID NOT NULL REFERENCES problems(id) ON DELETE CASCADE,
  problem_version_id UUID REFERENCES problem_versions(id) ON DELETE SET NULL,
  code TEXT NOT NULL,
  title TEXT NOT NULL,
  difficulty problem_difficulty NOT NULL,
  statement_excerpt TEXT NOT NULL DEFAULT '',
  limits_json JSONB NOT NULL DEFAULT '{}'::jsonb,
  metadata_json JSONB NOT NULL DEFAULT '{}'::jsonb,
  PRIMARY KEY (snapshot_id, code)
);

CREATE TABLE IF NOT EXISTS contest_public_summaries (
  contest_id UUID PRIMARY KEY REFERENCES contests(id) ON DELETE CASCADE,
  slug TEXT NOT NULL UNIQUE,
  title TEXT NOT NULL,
  status TEXT NOT NULL,
  starts_at TIMESTAMPTZ NOT NULL,
  ends_at TIMESTAMPTZ NOT NULL,
  duration_label TEXT NOT NULL,
  problem_count INTEGER NOT NULL DEFAULT 0,
  participant_count INTEGER NOT NULL DEFAULT 0,
  blurb TEXT NOT NULL DEFAULT '',
  updated_at TIMESTAMPTZ NOT NULL DEFAULT now()
);

CREATE TABLE IF NOT EXISTS contest_public_details (
  contest_id UUID PRIMARY KEY REFERENCES contests(id) ON DELETE CASCADE,
  slug TEXT NOT NULL UNIQUE,
  title TEXT NOT NULL,
  status TEXT NOT NULL,
  starts_at TIMESTAMPTZ NOT NULL,
  ends_at TIMESTAMPTZ NOT NULL,
  duration_label TEXT NOT NULL,
  problem_count INTEGER NOT NULL DEFAULT 0,
  participant_count INTEGER NOT NULL DEFAULT 0,
  blurb TEXT NOT NULL DEFAULT '',
  rank_summary TEXT NOT NULL DEFAULT '',
  remaining_label TEXT NOT NULL DEFAULT '',
  recent_submissions_json JSONB NOT NULL DEFAULT '[]'::jsonb,
  problems_json JSONB NOT NULL DEFAULT '[]'::jsonb,
  updated_at TIMESTAMPTZ NOT NULL DEFAULT now()
);

CREATE TABLE IF NOT EXISTS contest_scoreboard_rows (
  contest_id UUID NOT NULL REFERENCES contests(id) ON DELETE CASCADE,
  user_id UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
  rank INTEGER NOT NULL DEFAULT 0 CHECK (rank >= 0),
  solved_count INTEGER NOT NULL DEFAULT 0 CHECK (solved_count >= 0),
  penalty_seconds INTEGER NOT NULL DEFAULT 0 CHECK (penalty_seconds >= 0),
  row_json JSONB NOT NULL DEFAULT '{}'::jsonb,
  updated_at TIMESTAMPTZ NOT NULL DEFAULT now(),
  PRIMARY KEY (contest_id, user_id)
);

CREATE TABLE IF NOT EXISTS user_profiles (
  user_id UUID PRIMARY KEY REFERENCES users(id) ON DELETE CASCADE,
  display_name TEXT NOT NULL,
  preferred_locale TEXT NOT NULL DEFAULT 'zh',
  preferred_language TEXT NOT NULL DEFAULT 'CPP17',
  created_at TIMESTAMPTZ NOT NULL DEFAULT now(),
  updated_at TIMESTAMPTZ NOT NULL DEFAULT now()
);

CREATE TABLE IF NOT EXISTS user_stats (
  user_id UUID PRIMARY KEY REFERENCES users(id) ON DELETE CASCADE,
  solved_count INTEGER NOT NULL DEFAULT 0 CHECK (solved_count >= 0),
  submission_count INTEGER NOT NULL DEFAULT 0 CHECK (submission_count >= 0),
  accepted_count INTEGER NOT NULL DEFAULT 0 CHECK (accepted_count >= 0),
  last_active_at TIMESTAMPTZ,
  updated_at TIMESTAMPTZ NOT NULL DEFAULT now()
);

CREATE TABLE IF NOT EXISTS user_problem_statuses (
  user_id UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
  problem_id UUID NOT NULL REFERENCES problems(id) ON DELETE CASCADE,
  best_status submission_status NOT NULL DEFAULT 'PENDING',
  latest_status submission_status NOT NULL DEFAULT 'PENDING',
  first_accepted_at TIMESTAMPTZ,
  updated_at TIMESTAMPTZ NOT NULL DEFAULT now(),
  PRIMARY KEY (user_id, problem_id)
);

CREATE TABLE IF NOT EXISTS user_profile_snapshot (
  user_id UUID PRIMARY KEY REFERENCES users(id) ON DELETE CASCADE,
  summary_json JSONB NOT NULL DEFAULT '{}'::jsonb,
  updated_at TIMESTAMPTZ NOT NULL DEFAULT now()
);

CREATE TABLE IF NOT EXISTS user_recent_submission_snapshot (
  user_id UUID PRIMARY KEY REFERENCES users(id) ON DELETE CASCADE,
  submissions_json JSONB NOT NULL DEFAULT '[]'::jsonb,
  updated_at TIMESTAMPTZ NOT NULL DEFAULT now()
);

ALTER TABLE submissions
  ADD COLUMN IF NOT EXISTS contest_id UUID REFERENCES contests(id) ON DELETE SET NULL,
  ADD COLUMN IF NOT EXISTS result_snapshot_version INTEGER NOT NULL DEFAULT 0 CHECK (result_snapshot_version >= 0);

CREATE TABLE IF NOT EXISTS submission_artifacts (
  id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
  submission_id UUID NOT NULL REFERENCES submissions(id) ON DELETE CASCADE,
  artifact_type TEXT NOT NULL,
  object_key TEXT NOT NULL,
  content_type TEXT,
  content_encoding TEXT,
  expires_at TIMESTAMPTZ,
  created_at TIMESTAMPTZ NOT NULL DEFAULT now()
);

CREATE TABLE IF NOT EXISTS submission_summaries (
  submission_id UUID PRIMARY KEY REFERENCES submissions(id) ON DELETE CASCADE,
  user_id UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
  username TEXT NOT NULL,
  problem_id UUID NOT NULL REFERENCES problems(id) ON DELETE CASCADE,
  problem_json JSONB NOT NULL,
  language language NOT NULL,
  source_object_key TEXT NOT NULL,
  status submission_status NOT NULL,
  created_at TIMESTAMPTZ NOT NULL,
  updated_at TIMESTAMPTZ NOT NULL DEFAULT now()
);

CREATE TABLE IF NOT EXISTS submission_details (
  submission_id UUID PRIMARY KEY REFERENCES submissions(id) ON DELETE CASCADE,
  user_id UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
  username TEXT NOT NULL,
  problem_id UUID NOT NULL REFERENCES problems(id) ON DELETE CASCADE,
  problem_json JSONB NOT NULL,
  language language NOT NULL,
  source_object_key TEXT NOT NULL,
  status submission_status NOT NULL,
  created_at TIMESTAMPTZ NOT NULL,
  compile_summary JSONB NOT NULL DEFAULT '{}'::jsonb,
  results_json JSONB NOT NULL DEFAULT '[]'::jsonb,
  artifact_availability JSONB NOT NULL DEFAULT '{}'::jsonb,
  updated_at TIMESTAMPTZ NOT NULL DEFAULT now()
);

CREATE INDEX IF NOT EXISTS problem_public_summaries_updated_idx ON problem_public_summaries (updated_at DESC);
CREATE INDEX IF NOT EXISTS contest_public_summaries_status_starts_idx ON contest_public_summaries (status, starts_at DESC);
CREATE INDEX IF NOT EXISTS submission_summaries_user_created_idx ON submission_summaries (user_id, created_at DESC);
CREATE INDEX IF NOT EXISTS submission_summaries_username_created_idx ON submission_summaries (username, created_at DESC);
CREATE INDEX IF NOT EXISTS submission_details_problem_idx ON submission_details (problem_id, updated_at DESC);
CREATE INDEX IF NOT EXISTS user_problem_statuses_problem_idx ON user_problem_statuses (problem_id, updated_at DESC);
CREATE INDEX IF NOT EXISTS contest_scoreboard_rows_rank_idx ON contest_scoreboard_rows (contest_id, rank ASC);
CREATE INDEX IF NOT EXISTS submission_artifacts_submission_idx ON submission_artifacts (submission_id, expires_at);

INSERT INTO problem_versions (problem_id, version_no, manifest_object_key, public_snapshot_version, judge_bundle_version, status)
SELECT
  p.id,
  1,
  format('problems/%s/versions/1/manifest.json', p.id),
  1,
  1,
  CASE WHEN p.is_published THEN 'PUBLISHED' ELSE 'DRAFT' END
FROM problems p
ON CONFLICT (problem_id, version_no) DO NOTHING;

UPDATE problems p
SET
  current_public_version_id = pv.id,
  current_judge_version_id = pv.id
FROM problem_versions pv
WHERE pv.problem_id = p.id
  AND pv.version_no = 1
  AND (p.current_public_version_id IS NULL OR p.current_judge_version_id IS NULL);

INSERT INTO problem_stats (problem_id, submission_count, accepted_count, accepted_rate, updated_at)
SELECT
  p.id,
  COALESCE(stats.submission_count, 0),
  COALESCE(stats.accepted_count, 0),
  CASE
    WHEN COALESCE(stats.submission_count, 0) = 0 THEN 0
    ELSE ROUND((stats.accepted_count::numeric * 100) / stats.submission_count, 2)
  END,
  now()
FROM problems p
LEFT JOIN (
  SELECT
    s.problem_id,
    COUNT(*)::int AS submission_count,
    COUNT(*) FILTER (WHERE s.status = 'ACCEPTED')::int AS accepted_count
  FROM submissions s
  GROUP BY s.problem_id
) stats ON stats.problem_id = p.id
ON CONFLICT (problem_id) DO UPDATE SET
  submission_count = EXCLUDED.submission_count,
  accepted_count = EXCLUDED.accepted_count,
  accepted_rate = EXCLUDED.accepted_rate,
  updated_at = EXCLUDED.updated_at;

INSERT INTO problem_public_summaries (problem_id, slug, title, difficulty, tags_json, accepted_rate, statistics_json, updated_at)
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
ON CONFLICT (problem_id) DO UPDATE SET
  slug = EXCLUDED.slug,
  title = EXCLUDED.title,
  difficulty = EXCLUDED.difficulty,
  tags_json = EXCLUDED.tags_json,
  accepted_rate = EXCLUDED.accepted_rate,
  statistics_json = EXCLUDED.statistics_json,
  updated_at = EXCLUDED.updated_at;

INSERT INTO problem_public_details (problem_id, slug, title, difficulty, statement_json, samples_json, limits_json, metadata_json, updated_at)
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
ON CONFLICT (problem_id) DO UPDATE SET
  slug = EXCLUDED.slug,
  title = EXCLUDED.title,
  difficulty = EXCLUDED.difficulty,
  statement_json = EXCLUDED.statement_json,
  samples_json = EXCLUDED.samples_json,
  limits_json = EXCLUDED.limits_json,
  metadata_json = EXCLUDED.metadata_json,
  updated_at = EXCLUDED.updated_at;

INSERT INTO user_profiles (user_id, display_name, preferred_locale, preferred_language, updated_at)
SELECT
  u.id,
  u.username,
  'zh',
  'CPP17',
  now()
FROM users u
ON CONFLICT (user_id) DO UPDATE SET
  display_name = EXCLUDED.display_name,
  updated_at = EXCLUDED.updated_at;

INSERT INTO user_stats (user_id, solved_count, submission_count, accepted_count, last_active_at, updated_at)
SELECT
  u.id,
  COALESCE(accepted.solved_count, 0),
  COALESCE(activity.submission_count, 0),
  COALESCE(activity.accepted_count, 0),
  activity.last_active_at,
  now()
FROM users u
LEFT JOIN (
  SELECT
    s.user_id,
    COUNT(*)::int AS submission_count,
    COUNT(*) FILTER (WHERE s.status = 'ACCEPTED')::int AS accepted_count,
    MAX(s.created_at) AS last_active_at
  FROM submissions s
  GROUP BY s.user_id
) activity ON activity.user_id = u.id
LEFT JOIN (
  SELECT
    s.user_id,
    COUNT(DISTINCT s.problem_id)::int AS solved_count
  FROM submissions s
  WHERE s.status = 'ACCEPTED'
  GROUP BY s.user_id
) accepted ON accepted.user_id = u.id
ON CONFLICT (user_id) DO UPDATE SET
  solved_count = EXCLUDED.solved_count,
  submission_count = EXCLUDED.submission_count,
  accepted_count = EXCLUDED.accepted_count,
  last_active_at = EXCLUDED.last_active_at,
  updated_at = EXCLUDED.updated_at;

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
FROM (
  SELECT DISTINCT ON (s.user_id, s.problem_id)
    s.user_id,
    s.problem_id,
    s.status AS latest_status
  FROM submissions s
  ORDER BY s.user_id, s.problem_id, s.created_at DESC, s.id DESC
) latest
LEFT JOIN (
  SELECT
    s.user_id,
    s.problem_id,
    MIN(s.created_at) AS first_accepted_at
  FROM submissions s
  WHERE s.status = 'ACCEPTED'
  GROUP BY s.user_id, s.problem_id
) solved ON solved.user_id = latest.user_id AND solved.problem_id = latest.problem_id
ON CONFLICT (user_id, problem_id) DO UPDATE SET
  best_status = EXCLUDED.best_status,
  latest_status = EXCLUDED.latest_status,
  first_accepted_at = EXCLUDED.first_accepted_at,
  updated_at = EXCLUDED.updated_at;

INSERT INTO submission_summaries (submission_id, user_id, username, problem_id, problem_json, language, source_object_key, status, created_at, updated_at)
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
ON CONFLICT (submission_id) DO UPDATE SET
  user_id = EXCLUDED.user_id,
  username = EXCLUDED.username,
  problem_id = EXCLUDED.problem_id,
  problem_json = EXCLUDED.problem_json,
  language = EXCLUDED.language,
  source_object_key = EXCLUDED.source_object_key,
  status = EXCLUDED.status,
  created_at = EXCLUDED.created_at,
  updated_at = EXCLUDED.updated_at;

INSERT INTO submission_details (submission_id, user_id, username, problem_id, problem_json, language, source_object_key, status, created_at, compile_summary, results_json, artifact_availability, updated_at)
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
  updated_at = EXCLUDED.updated_at;

INSERT INTO user_profile_snapshot (user_id, summary_json, updated_at)
SELECT
  u.id,
  jsonb_build_object(
    'userId', u.id,
    'username', u.username,
    'displayName', up.display_name,
    'role', u.role,
    'status', u.status,
    'stats', jsonb_build_object(
      'solvedCount', COALESCE(us.solved_count, 0),
      'submissionCount', COALESCE(us.submission_count, 0),
      'acceptedCount', COALESCE(us.accepted_count, 0)
    )
  ),
  now()
FROM users u
LEFT JOIN user_profiles up ON up.user_id = u.id
LEFT JOIN user_stats us ON us.user_id = u.id
ON CONFLICT (user_id) DO UPDATE SET
  summary_json = EXCLUDED.summary_json,
  updated_at = EXCLUDED.updated_at;

INSERT INTO user_recent_submission_snapshot (user_id, submissions_json, updated_at)
SELECT
  u.id,
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
      WHERE ss.user_id = u.id
      ORDER BY ss.created_at DESC
      LIMIT 10
    ) item
  ), '[]'::jsonb),
  now()
FROM users u
ON CONFLICT (user_id) DO UPDATE SET
  submissions_json = EXCLUDED.submissions_json,
  updated_at = EXCLUDED.updated_at;

-- +goose Down
DROP TABLE IF EXISTS user_recent_submission_snapshot;
DROP TABLE IF EXISTS user_profile_snapshot;
DROP TABLE IF EXISTS submission_details;
DROP TABLE IF EXISTS submission_summaries;
DROP TABLE IF EXISTS submission_artifacts;
ALTER TABLE submissions
  DROP COLUMN IF EXISTS result_snapshot_version,
  DROP COLUMN IF EXISTS contest_id;
DROP TABLE IF EXISTS user_problem_statuses;
DROP TABLE IF EXISTS user_stats;
DROP TABLE IF EXISTS user_profiles;
DROP TABLE IF EXISTS contest_scoreboard_rows;
DROP TABLE IF EXISTS contest_public_details;
DROP TABLE IF EXISTS contest_public_summaries;
DROP TABLE IF EXISTS contest_problem_snapshots;
DROP TABLE IF EXISTS contest_snapshots;
DROP TABLE IF EXISTS contest_problem_links;
DROP TABLE IF EXISTS contests;
DROP TABLE IF EXISTS problem_public_details;
DROP TABLE IF EXISTS problem_public_summaries;
DROP TABLE IF EXISTS problem_stats;
DROP TABLE IF EXISTS problem_tag_links;
DROP TABLE IF EXISTS problem_tags;
ALTER TABLE problems
  DROP COLUMN IF EXISTS current_judge_version_id,
  DROP COLUMN IF EXISTS current_public_version_id;
DROP TABLE IF EXISTS problem_versions;
ALTER TABLE users
  DROP COLUMN IF EXISTS status;
