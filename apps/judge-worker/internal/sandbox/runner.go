package sandbox

import (
	"context"
	"os/exec"
	"strconv"
)

type Runner struct {
	Binary string
}

type RunSpec struct {
	Command       string
	Args          []string
	TimeLimitMs   int
	MemoryLimitKB int
	Workdir       string
}

func (r Runner) BuildArgs(spec RunSpec) []string {
	timeLimitSeconds := ceilMillisToSeconds(spec.TimeLimitMs)
	args := []string{
		"--disable_clone_newnet",
		"--quiet",
		"--time_limit",
		strconv.Itoa(timeLimitSeconds),
		"--rlimit_as",
		strconv.Itoa(spec.MemoryLimitKB),
		"--cwd",
		spec.Workdir,
		"--",
		spec.Command,
	}

	args = append(args, spec.Args...)
	return args
}

func (r Runner) Command(ctx context.Context, spec RunSpec) *exec.Cmd {
	binary := r.Binary
	if binary == "" {
		binary = "nsjail"
	}
	return exec.CommandContext(ctx, binary, r.BuildArgs(spec)...)
}

func ceilMillisToSeconds(ms int) int {
	if ms <= 0 {
		return 1
	}
	return (ms + 999) / 1000
}
