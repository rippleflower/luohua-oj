package sandbox_test

import (
	"testing"

	"github.com/example/oj3/apps/judge-worker/internal/sandbox"
	"github.com/stretchr/testify/require"
)

func TestBuildNSJailArgs(t *testing.T) {
	runner := sandbox.Runner{}
	args := runner.BuildArgs(sandbox.RunSpec{
		Command:       "/usr/bin/python3",
		Args:          []string{"main.py"},
		TimeLimitMs:   1000,
		MemoryLimitKB: 262144,
		Workdir:       "/tmp/oj-run-1",
	})

	require.Contains(t, args, "--disable_clone_newnet")
	require.Contains(t, args, "--time_limit")
	require.Contains(t, args, "1")
	require.Contains(t, args, "--rlimit_as")
	require.Contains(t, args, "262144")
	require.Contains(t, args, "/usr/bin/python3")
	require.Contains(t, args, "main.py")
}

func TestBuildNSJailArgsRoundsTimeLimitUp(t *testing.T) {
	runner := sandbox.Runner{}
	args := runner.BuildArgs(sandbox.RunSpec{
		Command:       "/bin/true",
		TimeLimitMs:   1500,
		MemoryLimitKB: 65536,
		Workdir:       "/tmp/oj-run-2",
	})

	require.Contains(t, args, "2")
}
