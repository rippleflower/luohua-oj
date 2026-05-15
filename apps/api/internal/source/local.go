package source

import (
	"context"
	"fmt"
	"os"
	"path/filepath"

	"github.com/google/uuid"
)

type LocalStore struct {
	Root string
}

func (s LocalStore) PutSource(ctx context.Context, submissionID uuid.UUID, source string) (string, error) {
	root := s.Root
	if root == "" {
		root = "tmp/submissions"
	}

	if err := os.MkdirAll(root, 0o755); err != nil {
		return "", err
	}

	object := filepath.Join(root, fmt.Sprintf("%s.txt", submissionID.String()))
	if err := os.WriteFile(object, []byte(source), 0o600); err != nil {
		return "", err
	}

	return object, nil
}
