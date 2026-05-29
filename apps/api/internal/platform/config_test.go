package platform

import (
	"os"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestLoadConfigIncludesProblemRouteSalt(t *testing.T) {
	t.Setenv("PROBLEM_ROUTE_SALT", "route-salt")
	t.Setenv("SESSION_DURATION_HOURS", "24")

	cfg := LoadConfig()

	require.Equal(t, "route-salt", cfg.ProblemRouteSalt)
	require.Equal(t, 24, cfg.SessionDurationHours)
}

func TestLoadConfigFallsBackOnInvalidInt(t *testing.T) {
	original := os.Getenv("SESSION_DURATION_HOURS")
	t.Setenv("SESSION_DURATION_HOURS", "invalid")
	defer os.Setenv("SESSION_DURATION_HOURS", original)

	cfg := LoadConfig()
	require.Equal(t, 336, cfg.SessionDurationHours)
}
