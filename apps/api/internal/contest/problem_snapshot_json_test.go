package contest

import (
	"encoding/json"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestProblemSnapshotJSONIncludesSlugWhenPresent(t *testing.T) {
	payload, err := json.Marshal([]ProblemSnapshot{
		{
			Code:       "A",
			Slug:       "two-sum",
			Title:      "Two Sum",
			Difficulty: "EASY",
			Status:     "LOCKED",
		},
	})
	require.NoError(t, err)
	require.Contains(t, string(payload), `"slug":"two-sum"`)
}

func TestProblemSnapshotJSONAcceptsLegacyShapeWithoutSlug(t *testing.T) {
	var snapshots []ProblemSnapshot
	err := json.Unmarshal(
		[]byte(`[{"code":"A","title":"Two Sum","difficulty":"EASY","status":"LOCKED"}]`),
		&snapshots,
	)
	require.NoError(t, err)
	require.Len(t, snapshots, 1)
	require.Equal(t, "", snapshots[0].Slug)
}
