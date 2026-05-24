-- +goose Up
ALTER TYPE user_role RENAME TO user_role_old;
CREATE TYPE user_role AS ENUM ('USER', 'ADMIN', 'SUPER_ADMIN');

ALTER TABLE users
  ALTER COLUMN role DROP DEFAULT,
  ALTER COLUMN role TYPE user_role
  USING (
    CASE
      WHEN role::text = 'PROBLEM_SETTER' THEN 'ADMIN'::user_role
      ELSE role::text::user_role
    END
  ),
  ALTER COLUMN role SET DEFAULT 'USER';

DROP TYPE user_role_old;

ALTER TABLE user_profiles
  ADD COLUMN IF NOT EXISTS bio TEXT NOT NULL DEFAULT '',
  ADD COLUMN IF NOT EXISTS avatar_url TEXT NOT NULL DEFAULT '';

INSERT INTO user_profiles (user_id, display_name, preferred_locale, preferred_language, bio, avatar_url)
SELECT
  u.id,
  u.username,
  'zh',
  'CPP17',
  '',
  ''
FROM users u
LEFT JOIN user_profiles up ON up.user_id = u.id
WHERE up.user_id IS NULL;

CREATE TABLE user_sessions (
  id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
  user_id UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
  session_token_hash TEXT NOT NULL UNIQUE,
  ip TEXT NOT NULL DEFAULT '',
  user_agent TEXT NOT NULL DEFAULT '',
  expires_at TIMESTAMPTZ NOT NULL,
  last_seen_at TIMESTAMPTZ,
  created_at TIMESTAMPTZ NOT NULL DEFAULT now()
);

CREATE TABLE admin_permission_grants (
  user_id UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
  permission_key TEXT NOT NULL,
  granted_by UUID REFERENCES users(id) ON DELETE SET NULL,
  created_at TIMESTAMPTZ NOT NULL DEFAULT now(),
  PRIMARY KEY (user_id, permission_key)
);

CREATE TABLE audit_logs (
  id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
  actor_user_id UUID REFERENCES users(id) ON DELETE SET NULL,
  actor_role user_role,
  action TEXT NOT NULL,
  target_type TEXT NOT NULL,
  target_id TEXT NOT NULL,
  diff_json JSONB NOT NULL DEFAULT '{}'::jsonb,
  reason TEXT NOT NULL DEFAULT '',
  ip TEXT NOT NULL DEFAULT '',
  created_at TIMESTAMPTZ NOT NULL DEFAULT now()
);

CREATE TABLE announcements (
  id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
  title TEXT NOT NULL,
  content TEXT NOT NULL,
  status TEXT NOT NULL DEFAULT 'DRAFT',
  audience TEXT NOT NULL DEFAULT 'ALL',
  created_by UUID REFERENCES users(id) ON DELETE SET NULL,
  created_at TIMESTAMPTZ NOT NULL DEFAULT now(),
  updated_at TIMESTAMPTZ NOT NULL DEFAULT now()
);

CREATE TABLE system_settings (
  singleton BOOLEAN PRIMARY KEY DEFAULT true CHECK (singleton),
  settings_json JSONB NOT NULL DEFAULT '{}'::jsonb,
  updated_by UUID REFERENCES users(id) ON DELETE SET NULL,
  updated_at TIMESTAMPTZ NOT NULL DEFAULT now()
);

INSERT INTO system_settings (singleton, settings_json)
VALUES (
  true,
  jsonb_build_object(
    'registrationEnabled', true,
    'judgeQueuePaused', false,
    'storageMode', 'LOCAL'
  )
)
ON CONFLICT (singleton) DO NOTHING;

INSERT INTO admin_permission_grants (user_id, permission_key)
SELECT u.id, permissions.permission_key
FROM users u
CROSS JOIN (
  VALUES
    ('dashboard.view'),
    ('users.view'),
    ('users.edit'),
    ('users.roles'),
    ('problems.view'),
    ('problems.edit'),
    ('problems.publish'),
    ('contests.view'),
    ('contests.edit'),
    ('contests.publish'),
    ('submissions.view'),
    ('submissions.rejudge'),
    ('announcements.view'),
    ('announcements.edit'),
    ('system.view'),
    ('system.edit'),
    ('audit.view')
) AS permissions(permission_key)
WHERE u.role = 'ADMIN'
ON CONFLICT (user_id, permission_key) DO NOTHING;

CREATE INDEX user_sessions_user_expires_idx ON user_sessions (user_id, expires_at DESC);
CREATE INDEX user_sessions_hash_idx ON user_sessions (session_token_hash);
CREATE INDEX admin_permission_grants_user_idx ON admin_permission_grants (user_id);
CREATE INDEX audit_logs_actor_created_idx ON audit_logs (actor_user_id, created_at DESC);
CREATE INDEX audit_logs_target_idx ON audit_logs (target_type, target_id, created_at DESC);
CREATE INDEX announcements_status_updated_idx ON announcements (status, updated_at DESC);

-- +goose Down
DROP INDEX IF EXISTS announcements_status_updated_idx;
DROP INDEX IF EXISTS audit_logs_target_idx;
DROP INDEX IF EXISTS audit_logs_actor_created_idx;
DROP INDEX IF EXISTS admin_permission_grants_user_idx;
DROP INDEX IF EXISTS user_sessions_hash_idx;
DROP INDEX IF EXISTS user_sessions_user_expires_idx;
DROP TABLE IF EXISTS system_settings;
DROP TABLE IF EXISTS announcements;
DROP TABLE IF EXISTS audit_logs;
DROP TABLE IF EXISTS admin_permission_grants;
DROP TABLE IF EXISTS user_sessions;

ALTER TABLE user_profiles
  DROP COLUMN IF EXISTS avatar_url,
  DROP COLUMN IF EXISTS bio;

ALTER TYPE user_role RENAME TO user_role_new;
CREATE TYPE user_role AS ENUM ('USER', 'PROBLEM_SETTER', 'ADMIN');

ALTER TABLE users
  ALTER COLUMN role DROP DEFAULT,
  ALTER COLUMN role TYPE user_role
  USING (
    CASE
      WHEN role::text = 'SUPER_ADMIN' THEN 'ADMIN'::user_role
      ELSE role::text::user_role
    END
  ),
  ALTER COLUMN role SET DEFAULT 'USER';

DROP TYPE user_role_new;
