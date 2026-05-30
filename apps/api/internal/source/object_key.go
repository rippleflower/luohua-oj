package source

import (
	"fmt"
	"strings"
	"time"

	"github.com/google/uuid"
)

func SubmissionSourceObjectKey(now time.Time, submissionID uuid.UUID, language string) string {
	timestamp := now.UTC()
	return fmt.Sprintf(
		"submissions/%04d/%02d/%s/source.%s.zst",
		timestamp.Year(),
		timestamp.Month(),
		submissionID.String(),
		languageExtension(language),
	)
}

func languageExtension(language string) string {
	switch strings.ToUpper(strings.TrimSpace(language)) {
	case "CPP17", "CPP20":
		return "cpp"
	case "JAVA17":
		return "java"
	case "PYTHON311":
		return "py"
	default:
		return "txt"
	}
}
