package judge

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"io"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"syscall"
	"time"

	"github.com/example/oj3/apps/judge-worker/internal/sandbox"
)

const snippetLimit = 2048

type commandRunner interface {
	Command(ctx context.Context, spec sandbox.RunSpec) *exec.Cmd
}

type LocalExecutor struct {
	SourceRoot string
	WorkRoot   string
	Runner     commandRunner
}

type JudgeOutcome struct {
	Status        Status
	CompileOutput string
	Results       []RunResult
}

type languageConfig struct {
	sourceFileName string
	compileCommand []string
	runCommand     []string
}

func (e LocalExecutor) Judge(ctx context.Context, submission Submission, testCases []TestCase) (JudgeOutcome, error) {
	language, err := languageConfigFor(submission.Language)
	if err != nil {
		return JudgeOutcome{}, err
	}

	workRoot := e.WorkRoot
	if strings.TrimSpace(workRoot) == "" {
		workRoot = "tmp/judge-runs"
	}

	workdir, err := os.MkdirTemp(workRoot, submission.ID.String()+"-")
	if err != nil {
		return JudgeOutcome{}, err
	}
	defer os.RemoveAll(workdir)

	if err := os.MkdirAll(workdir, 0o755); err != nil {
		return JudgeOutcome{}, err
	}

	sourcePath := filepath.Join(workdir, language.sourceFileName)
	if err := copyObjectToPath(e.SourceRoot, submission.SourceObject, sourcePath); err != nil {
		return JudgeOutcome{}, fmt.Errorf("load source: %w", err)
	}

	runner := e.Runner
	if runner == nil {
		runner = sandbox.Runner{}
	}

	compileOutput, compileStatus, err := compileProgram(ctx, runner, workdir, submission, language)
	if err != nil {
		return JudgeOutcome{}, err
	}
	if compileStatus != "" {
		return JudgeOutcome{
			Status:        compileStatus,
			CompileOutput: compileOutput,
		}, nil
	}

	results := make([]RunResult, 0, len(testCases))
	for _, testCase := range testCases {
		inputData, err := readObject(e.SourceRoot, testCase.InputObject)
		if err != nil {
			return JudgeOutcome{}, fmt.Errorf("load test case input: %w", err)
		}
		expectedOutput, err := readObject(e.SourceRoot, testCase.OutputObject)
		if err != nil {
			return JudgeOutcome{}, fmt.Errorf("load test case output: %w", err)
		}

		result, err := runTestCase(ctx, runner, workdir, submission, language, inputData, expectedOutput)
		if err != nil {
			return JudgeOutcome{}, err
		}
		results = append(results, result)
		if result.Status != StatusAccepted {
			return JudgeOutcome{
				Status:        result.Status,
				CompileOutput: compileOutput,
				Results:       results,
			}, nil
		}
	}

	return JudgeOutcome{
		Status:        StatusAccepted,
		CompileOutput: compileOutput,
		Results:       results,
	}, nil
}

func compileProgram(ctx context.Context, runner commandRunner, workdir string, submission Submission, language languageConfig) (string, Status, error) {
	if len(language.compileCommand) == 0 {
		return "", "", nil
	}

	result, err := runCommand(
		ctx,
		runner,
		sandbox.RunSpec{
			Command:       language.compileCommand[0],
			Args:          language.compileCommand[1:],
			TimeLimitMs:   compileTimeLimitMs(submission.TimeLimitMs),
			MemoryLimitKB: compileMemoryLimitKB(submission.MemoryLimitKB),
			Workdir:       workdir,
		},
		nil,
	)
	if err != nil {
		return "", "", err
	}

	output := truncateSnippet(strings.TrimSpace(result.Stderr + "\n" + result.Stdout))
	if result.TimedOut || result.ExitCode != 0 {
		if output == "" {
			output = "compile failed"
		}
		return output, StatusCompileError, nil
	}

	return output, "", nil
}

func runTestCase(ctx context.Context, runner commandRunner, workdir string, submission Submission, language languageConfig, inputData []byte, expectedOutput []byte) (RunResult, error) {
	result, err := runCommand(
		ctx,
		runner,
		sandbox.RunSpec{
			Command:       language.runCommand[0],
			Args:          language.runCommand[1:],
			TimeLimitMs:   int(submission.TimeLimitMs),
			MemoryLimitKB: int(submission.MemoryLimitKB),
			Workdir:       workdir,
		},
		inputData,
	)
	if err != nil {
		return RunResult{}, err
	}

	outputSnippet := truncateSnippet(result.Stdout)
	errorSnippet := truncateSnippet(result.Stderr)
	switch {
	case result.TimedOut:
		return RunResult{
			Status:   StatusTimeLimitExceeded,
			Output:   outputSnippet,
			Error:    errorSnippet,
			TimeMs:   result.TimeMs,
			MemoryKB: result.MemoryKB,
		}, nil
	case isMemoryLimitError(result.Stderr):
		return RunResult{
			Status:   StatusMemoryLimitExceeded,
			Output:   outputSnippet,
			Error:    errorSnippet,
			TimeMs:   result.TimeMs,
			MemoryKB: result.MemoryKB,
		}, nil
	case result.ExitCode != 0:
		return RunResult{
			Status:   StatusRuntimeError,
			Output:   outputSnippet,
			Error:    errorSnippet,
			TimeMs:   result.TimeMs,
			MemoryKB: result.MemoryKB,
		}, nil
	case !bytes.Equal(result.StdoutBytes, expectedOutput):
		return RunResult{
			Status:   StatusWrongAnswer,
			Output:   outputSnippet,
			TimeMs:   result.TimeMs,
			MemoryKB: result.MemoryKB,
		}, nil
	default:
		return RunResult{
			Status:   StatusAccepted,
			Output:   outputSnippet,
			TimeMs:   result.TimeMs,
			MemoryKB: result.MemoryKB,
		}, nil
	}
}

type commandResult struct {
	Stdout      string
	Stderr      string
	StdoutBytes []byte
	TimeMs      int32
	MemoryKB    int32
	ExitCode    int
	TimedOut    bool
}

func runCommand(ctx context.Context, runner commandRunner, spec sandbox.RunSpec, stdin []byte) (commandResult, error) {
	limit := time.Duration(spec.TimeLimitMs+1000) * time.Millisecond
	if spec.TimeLimitMs <= 0 {
		limit = 2 * time.Second
	}
	commandCtx, cancel := context.WithTimeout(ctx, limit)
	defer cancel()

	cmd := runner.Command(commandCtx, spec)
	var stdout bytes.Buffer
	var stderr bytes.Buffer
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr
	if stdin != nil {
		cmd.Stdin = bytes.NewReader(stdin)
	}

	startedAt := time.Now()
	err := cmd.Run()
	result := commandResult{
		Stdout:      stdout.String(),
		Stderr:      stderr.String(),
		StdoutBytes: stdout.Bytes(),
		TimeMs:      elapsedMillis(time.Since(startedAt)),
		MemoryKB:    maxRSS(cmd.ProcessState),
	}

	if errors.Is(commandCtx.Err(), context.DeadlineExceeded) {
		result.TimedOut = true
	}

	if err == nil {
		return result, nil
	}

	var exitErr *exec.ExitError
	if errors.As(err, &exitErr) {
		result.ExitCode = exitErr.ExitCode()
		return result, nil
	}

	return result, err
}

func languageConfigFor(language string) (languageConfig, error) {
	switch strings.ToUpper(strings.TrimSpace(language)) {
	case "CPP17":
		return languageConfig{
			sourceFileName: "main.cpp",
			compileCommand: []string{"g++", "-std=c++17", "-O2", "-pipe", "main.cpp", "-o", "main"},
			runCommand:     []string{"./main"},
		}, nil
	case "CPP20":
		return languageConfig{
			sourceFileName: "main.cpp",
			compileCommand: []string{"g++", "-std=c++20", "-O2", "-pipe", "main.cpp", "-o", "main"},
			runCommand:     []string{"./main"},
		}, nil
	case "JAVA17":
		return languageConfig{
			sourceFileName: "Main.java",
			compileCommand: []string{"javac", "Main.java"},
			runCommand:     []string{"java", "-cp", ".", "Main"},
		}, nil
	case "PYTHON311":
		return languageConfig{
			sourceFileName: "main.py",
			compileCommand: []string{"python3", "-m", "py_compile", "main.py"},
			runCommand:     []string{"python3", "main.py"},
		}, nil
	default:
		return languageConfig{}, fmt.Errorf("unsupported language: %s", language)
	}
}

func copyObjectToPath(root string, objectKey string, destination string) error {
	content, err := readObject(root, objectKey)
	if err != nil {
		return err
	}
	return os.WriteFile(destination, content, 0o600)
}

func readObject(root string, objectKey string) ([]byte, error) {
	path := resolveObjectPath(root, objectKey)
	file, err := os.Open(path)
	if err != nil {
		return nil, err
	}
	defer file.Close()

	if strings.HasSuffix(strings.ToLower(path), ".zst") {
		cmd := exec.Command("zstd", "-d", "-q", "-c", path)
		return cmd.Output()
	}

	return io.ReadAll(file)
}

func resolveObjectPath(root string, objectKey string) string {
	if filepath.IsAbs(objectKey) {
		return objectKey
	}

	base := root
	if strings.TrimSpace(base) == "" {
		base = "tmp/submissions"
	}
	return filepath.Join(base, filepath.FromSlash(objectKey))
}

func compileTimeLimitMs(runLimit int32) int {
	if runLimit < 10_000 {
		return 10_000
	}
	return int(runLimit)
}

func compileMemoryLimitKB(runLimit int32) int {
	if runLimit < 1_048_576 {
		return 1_048_576
	}
	return int(runLimit)
}

func elapsedMillis(duration time.Duration) int32 {
	if duration <= 0 {
		return 0
	}
	if duration < time.Millisecond {
		return 1
	}
	return int32(duration.Milliseconds())
}

func maxRSS(state *os.ProcessState) int32 {
	if state == nil {
		return 0
	}

	usage, ok := state.SysUsage().(*syscall.Rusage)
	if !ok || usage == nil {
		return 0
	}

	if usage.Maxrss <= 0 {
		return 0
	}

	return int32(usage.Maxrss)
}

func isMemoryLimitError(stderr string) bool {
	lowered := strings.ToLower(stderr)
	return strings.Contains(lowered, "memory") && strings.Contains(lowered, "limit")
}

func truncateSnippet(value string) string {
	if len(value) <= snippetLimit {
		return value
	}
	return value[:snippetLimit]
}
