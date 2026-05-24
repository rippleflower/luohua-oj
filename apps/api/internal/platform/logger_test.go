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
		LogFileName: "api.log",
		LogFormat:   "json",
		LogLevel:    "debug",
	}, "api")
	require.NoError(t, err)
	t.Cleanup(func() {
		require.NoError(t, closer.Close())
	})

	logger.Info("submission accepted", "component", "submission", "event", "submission.created")

	content, err := os.ReadFile(filepath.Join(dir, "api.log"))
	require.NoError(t, err)
	require.Contains(t, string(content), "\"service\":\"api\"")
	require.Contains(t, string(content), "\"event\":\"submission.created\"")
	require.True(t, strings.Contains(string(content), "\"component\":\"submission\""))
}
