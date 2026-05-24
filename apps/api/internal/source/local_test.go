package source_test

import (
	"context"
	"os/exec"
	"testing"
	"time"

	"github.com/example/oj3/apps/api/internal/source"
	"github.com/google/uuid"
	"github.com/stretchr/testify/require"
)

func TestLocalStorePutSourceUsesStableObjectKey(t *testing.T) {
	root := t.TempDir()
	store := source.LocalStore{
		Root: root,
		Clock: func() time.Time {
			return time.Date(2026, 5, 16, 9, 30, 0, 0, time.UTC)
		},
	}
	submissionID := uuid.MustParse("aaaaaaaa-bbbb-cccc-dddd-eeeeeeeeeeee")

	objectKey, err := store.PutSource(context.Background(), submissionID, "CPP17", "int main(){return 0;}")

	require.NoError(t, err)
	require.Equal(t, "submissions/2026/05/aaaaaaaa-bbbb-cccc-dddd-eeeeeeeeeeee/source.cpp.zst", objectKey)

	content, err := exec.Command("zstd", "-d", "-q", "-c", store.ResolvePath(objectKey)).Output()
	require.NoError(t, err)
	require.Equal(t, "int main(){return 0;}", string(content))
}
