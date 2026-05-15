-- +goose Up
CREATE EXTENSION IF NOT EXISTS pgcrypto;

CREATE TYPE user_role AS ENUM ('USER', 'PROBLEM_SETTER', 'ADMIN');
CREATE TYPE problem_difficulty AS ENUM ('EASY', 'MEDIUM', 'HARD');
CREATE TYPE language AS ENUM ('CPP17', 'CPP20', 'JAVA17', 'PYTHON311');
CREATE TYPE submission_status AS ENUM (
  'PENDING',
  'RUNNING',
  'ACCEPTED',
  'WRONG_ANSWER',
  'TIME_LIMIT_EXCEEDED',
  'MEMORY_LIMIT_EXCEEDED',
  'RUNTIME_ERROR',
  'COMPILE_ERROR',
  'SYSTEM_ERROR'
);

CREATE TABLE users (
  id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
  email TEXT NOT NULL UNIQUE,
  username TEXT NOT NULL UNIQUE,
  password_hash TEXT NOT NULL,
  role user_role NOT NULL DEFAULT 'USER',
  created_at TIMESTAMPTZ NOT NULL DEFAULT now(),
  updated_at TIMESTAMPTZ NOT NULL DEFAULT now()
);

CREATE TABLE problems (
  id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
  slug TEXT NOT NULL UNIQUE,
  title TEXT NOT NULL,
  difficulty problem_difficulty NOT NULL,
  statement_md TEXT NOT NULL,
  input_md TEXT NOT NULL,
  output_md TEXT NOT NULL,
  constraints_md TEXT NOT NULL,
  time_limit_ms INTEGER NOT NULL CHECK (time_limit_ms > 0),
  memory_limit_kb INTEGER NOT NULL CHECK (memory_limit_kb > 0),
  is_published BOOLEAN NOT NULL DEFAULT false,
  created_at TIMESTAMPTZ NOT NULL DEFAULT now(),
  updated_at TIMESTAMPTZ NOT NULL DEFAULT now()
);

CREATE TABLE test_cases (
  id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
  problem_id UUID NOT NULL REFERENCES problems(id) ON DELETE CASCADE,
  input_object TEXT NOT NULL,
  output_object TEXT NOT NULL,
  is_sample BOOLEAN NOT NULL DEFAULT false,
  weight INTEGER NOT NULL DEFAULT 1 CHECK (weight > 0),
  created_at TIMESTAMPTZ NOT NULL DEFAULT now()
);

CREATE TABLE submissions (
  id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
  user_id UUID NOT NULL REFERENCES users(id),
  problem_id UUID NOT NULL REFERENCES problems(id),
  language language NOT NULL,
  source_object TEXT NOT NULL,
  status submission_status NOT NULL DEFAULT 'PENDING',
  score INTEGER NOT NULL DEFAULT 0 CHECK (score >= 0),
  compile_output TEXT,
  max_time_ms INTEGER,
  max_memory_kb INTEGER,
  created_at TIMESTAMPTZ NOT NULL DEFAULT now(),
  judged_at TIMESTAMPTZ
);

CREATE TABLE submission_results (
  id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
  submission_id UUID NOT NULL REFERENCES submissions(id) ON DELETE CASCADE,
  test_case_id UUID NOT NULL REFERENCES test_cases(id),
  status submission_status NOT NULL,
  time_ms INTEGER,
  memory_kb INTEGER,
  output_snippet TEXT,
  error_snippet TEXT
);

CREATE INDEX submissions_user_created_idx ON submissions (user_id, created_at DESC);
CREATE INDEX submissions_problem_created_idx ON submissions (problem_id, created_at DESC);
CREATE INDEX submission_results_submission_idx ON submission_results (submission_id);
CREATE INDEX test_cases_problem_idx ON test_cases (problem_id);

-- +goose Down
DROP TABLE IF EXISTS submission_results;
DROP TABLE IF EXISTS submissions;
DROP TABLE IF EXISTS test_cases;
DROP TABLE IF EXISTS problems;
DROP TABLE IF EXISTS users;
DROP TYPE IF EXISTS submission_status;
DROP TYPE IF EXISTS language;
DROP TYPE IF EXISTS problem_difficulty;
DROP TYPE IF EXISTS user_role;
