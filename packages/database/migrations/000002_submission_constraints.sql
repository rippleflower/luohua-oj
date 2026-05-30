-- +goose Up
ALTER TABLE submissions
  ADD CONSTRAINT submissions_max_time_ms_nonnegative CHECK (max_time_ms IS NULL OR max_time_ms >= 0),
  ADD CONSTRAINT submissions_max_memory_kb_nonnegative CHECK (max_memory_kb IS NULL OR max_memory_kb >= 0);

ALTER TABLE submission_results
  ADD CONSTRAINT submission_results_time_ms_nonnegative CHECK (time_ms IS NULL OR time_ms >= 0),
  ADD CONSTRAINT submission_results_memory_kb_nonnegative CHECK (memory_kb IS NULL OR memory_kb >= 0),
  ADD CONSTRAINT submission_results_submission_test_case_unique UNIQUE (submission_id, test_case_id);

-- +goose Down
ALTER TABLE submission_results
  DROP CONSTRAINT IF EXISTS submission_results_submission_test_case_unique,
  DROP CONSTRAINT IF EXISTS submission_results_memory_kb_nonnegative,
  DROP CONSTRAINT IF EXISTS submission_results_time_ms_nonnegative;

ALTER TABLE submissions
  DROP CONSTRAINT IF EXISTS submissions_max_memory_kb_nonnegative,
  DROP CONSTRAINT IF EXISTS submissions_max_time_ms_nonnegative;
