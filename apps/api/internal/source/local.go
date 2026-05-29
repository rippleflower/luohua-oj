package source

import (
	"context"
	"io"
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

	compressed, err := compressZstd(ctx, strings.NewReader(source))
	if err != nil {
		return "", err
	}

	if err := os.WriteFile(path, compressed, 0o600); err != nil {
		return "", err
	}

	return objectKey, nil
}

func (s LocalStore) PutObject(_ context.Context, objectKey string, content []byte) error {
	path := s.ResolvePath(objectKey)
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		return err
	}
	return os.WriteFile(path, content, 0o600)
}

func (s LocalStore) ReadObject(objectKey string) ([]byte, error) {
	path := s.ResolvePath(objectKey)
	file, err := os.Open(path)
	if err != nil {
		return nil, err
	}
	defer file.Close()

	if strings.HasSuffix(strings.ToLower(path), ".zst") {
		return exec.Command("zstd", "-d", "-q", "-c", path).Output()
	}

	return io.ReadAll(file)
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

func compressZstd(ctx context.Context, reader io.Reader) ([]byte, error) {
	cmd := exec.CommandContext(ctx, "zstd", "-q", "--stdout")
	cmd.Stdin = reader
	return cmd.Output()
}
