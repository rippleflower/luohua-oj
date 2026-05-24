package platform

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestNewLoggerWritesToFile(t *testing.T) {
	dir := t.TempDir()

	logger, closer, err := NewLogger(Config{
		LogDir:      dir,
		LogFileName: "judge-worker.log",
		LogFormat:   "json",
		LogLevel:    "debug",
	}, "judge-worker")
	require.NoError(t, err)
	t.Cleanup(func() {
		require.NoError(t, closer.Close())
	})

	logger.Info("task completed", "component", "judge.processor", "event", "judge.task.completed")

	content, err := os.ReadFile(filepath.Join(dir, "judge-worker.log"))
	require.NoError(t, err)
	require.Contains(t, string(content), "\"service\":\"judge-worker\"")
	require.Contains(t, string(content), "\"event\":\"judge.task.completed\"")
	require.True(t, strings.Contains(string(content), "\"component\":\"judge.processor\""))
}
