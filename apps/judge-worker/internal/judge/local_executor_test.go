package judge

import (
	"context"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/example/oj3/apps/judge-worker/internal/sandbox"
	"github.com/google/uuid"
	"github.com/stretchr/testify/require"
)

type directRunner struct{}

func (directRunner) Command(ctx context.Context, spec sandbox.RunSpec) *exec.Cmd {
	cmd := exec.CommandContext(ctx, spec.Command, spec.Args...)
	cmd.Dir = spec.Workdir
	return cmd
}

func TestLanguageConfigForPython(t *testing.T) {
	config, err := languageConfigFor("PYTHON311")

	require.NoError(t, err)
	require.Equal(t, "main.py", config.sourceFileName)
	require.Equal(t, []string{"python3", "-m", "py_compile", "main.py"}, config.compileCommand)
	require.Equal(t, []string{"python3", "main.py"}, config.runCommand)
}

func TestReadObjectSupportsCompressedAndPlainFiles(t *testing.T) {
	root := t.TempDir()
	plainPath := filepath.Join(root, "cases", "sample.in")
	require.NoError(t, os.MkdirAll(filepath.Dir(plainPath), 0o755))
	require.NoError(t, os.WriteFile(plainPath, []byte("42\n"), 0o600))

	compressedPath := filepath.Join(root, "submissions", "sample.py.zst")
	require.NoError(t, os.MkdirAll(filepath.Dir(compressedPath), 0o755))
	cmd := exec.Command("zstd", "-q", "-f", "-o", compressedPath)
	cmd.Stdin = strings.NewReader("print(input())\n")
	require.NoError(t, cmd.Run())

	plainContent, err := readObject(root, "cases/sample.in")
	require.NoError(t, err)
	require.Equal(t, []byte("42\n"), plainContent)

	compressedContent, err := readObject(root, "submissions/sample.py.zst")
	require.NoError(t, err)
	require.Equal(t, []byte("print(input())\n"), compressedContent)
}

func TestLocalExecutorJudgesPythonSubmission(t *testing.T) {
	if _, err := exec.LookPath("python3"); err != nil {
		t.Skip("python3 is required for local executor test")
	}

	root := t.TempDir()
	workRoot := t.TempDir()
	submissionID := uuid.New()
	sourceKey := "submissions/2026/05/" + submissionID.String() + "/source.py.zst"
	sourcePath := filepath.Join(root, filepath.FromSlash(sourceKey))
	require.NoError(t, os.MkdirAll(filepath.Dir(sourcePath), 0o755))
	compress := exec.Command("zstd", "-q", "-f", "-o", sourcePath)
	compress.Stdin = strings.NewReader("print(input().strip())\n")
	require.NoError(t, compress.Run())

	inputKey := "cases/demo.in"
	outputKey := "cases/demo.out"
	require.NoError(t, os.MkdirAll(filepath.Join(root, "cases"), 0o755))
	require.NoError(t, os.WriteFile(filepath.Join(root, "cases", "demo.in"), []byte("hello\n"), 0o600))
	require.NoError(t, os.WriteFile(filepath.Join(root, "cases", "demo.out"), []byte("hello\n"), 0o600))

	executor := LocalExecutor{
		SourceRoot: root,
		WorkRoot:   workRoot,
		Runner:     directRunner{},
	}

	outcome, err := executor.Judge(context.Background(), Submission{
		ID:            submissionID,
		ProblemID:     uuid.New(),
		Language:      "PYTHON311",
		SourceObject:  sourceKey,
		TimeLimitMs:   1000,
		MemoryLimitKB: 262144,
	}, []TestCase{{ID: uuid.New(), InputObject: inputKey, OutputObject: outputKey}})

	require.NoError(t, err)
	require.Equal(t, StatusAccepted, outcome.Status)
	require.Len(t, outcome.Results, 1)
	require.Equal(t, StatusAccepted, outcome.Results[0].Status)
	require.GreaterOrEqual(t, outcome.Results[0].TimeMs, int32(1))
}

func TestLocalExecutorReturnsCompileErrorForInvalidPython(t *testing.T) {
	if _, err := exec.LookPath("python3"); err != nil {
		t.Skip("python3 is required for local executor test")
	}

	root := t.TempDir()
	workRoot := t.TempDir()
	submissionID := uuid.New()
	sourceKey := "submissions/2026/05/" + submissionID.String() + "/source.py.zst"
	sourcePath := filepath.Join(root, filepath.FromSlash(sourceKey))
	require.NoError(t, os.MkdirAll(filepath.Dir(sourcePath), 0o755))
	compress := exec.Command("zstd", "-q", "-f", "-o", sourcePath)
	compress.Stdin = strings.NewReader("print(\n")
	require.NoError(t, compress.Run())

	executor := LocalExecutor{
		SourceRoot: root,
		WorkRoot:   workRoot,
		Runner:     directRunner{},
	}

	outcome, err := executor.Judge(context.Background(), Submission{
		ID:            submissionID,
		ProblemID:     uuid.New(),
		Language:      "PYTHON311",
		SourceObject:  sourceKey,
		TimeLimitMs:   1000,
		MemoryLimitKB: 262144,
	}, []TestCase{{ID: uuid.New(), InputObject: "cases/unused.in", OutputObject: "cases/unused.out"}})

	require.NoError(t, err)
	require.Equal(t, StatusCompileError, outcome.Status)
	require.Contains(t, outcome.CompileOutput, "SyntaxError")
}

func TestCompileHelpersFloorToSafeMinimums(t *testing.T) {
	require.Equal(t, 10_000, compileTimeLimitMs(500))
	require.Equal(t, 1_048_576, compileMemoryLimitKB(262144))
}

func TestElapsedMillisRoundsSubMillisecondDurationsUp(t *testing.T) {
	require.Equal(t, int32(1), elapsedMillis(500*time.Microsecond))
}
