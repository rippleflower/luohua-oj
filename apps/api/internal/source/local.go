package source

import (
	"context"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"time"

	"github.com/google/uuid"
)

type LocalStore struct {
	Root  string
	Clock func() time.Time
}

func (s LocalStore) PutSource(ctx context.Context, submissionID uuid.UUID, language string, source string) (string, error) {
	root := s.Root
	if root == "" {
		root = "tmp/submissions"
	}

	objectKey := SubmissionSourceObjectKey(s.now(), submissionID, language)
	path := filepath.Join(root, filepath.FromSlash(objectKey))
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		return "", err
	}

	cmd := exec.CommandContext(ctx, "zstd", "-q", "--stdout")
	cmd.Stdin = strings.NewReader(source)
	compressed, err := cmd.Output()
	if err != nil {
		return "", err
	}

	if err := os.WriteFile(path, compressed, 0o600); err != nil {
		return "", err
	}

	return objectKey, nil
}

func (s LocalStore) ResolvePath(objectKey string) string {
	root := s.Root
	if root == "" {
		root = "tmp/submissions"
	}
	return filepath.Join(root, filepath.FromSlash(objectKey))
}

func (s LocalStore) now() time.Time {
	if s.Clock != nil {
		return s.Clock()
	}
	return time.Now().UTC()
}
