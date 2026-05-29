package platform

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestLoadConfigIncludesWorkerConcurrency(t *testing.T) {
	t.Setenv("JUDGE_WORKER_CONCURRENCY", "6")
	t.Setenv("PROBLEM_ROUTE_SALT", "route-salt")

	cfg := LoadConfig()

	require.Equal(t, 6, cfg.WorkerConcurrency)
	require.Equal(t, "route-salt", cfg.ProblemRouteSalt)
}

func TestLoadConfigFallsBackOnInvalidWorkerConcurrency(t *testing.T) {
	t.Setenv("JUDGE_WORKER_CONCURRENCY", "invalid")

	cfg := LoadConfig()

	require.Equal(t, 4, cfg.WorkerConcurrency)
}
