package contest_makeup

import (
	"context"
	"errors"
	"time"

	"github.com/example/oj3/apps/api/internal/auth"
	"github.com/google/uuid"
)

var (
	ErrContestNotFound = errors.New("contest not found")
	ErrContestNotEnded = errors.New("contest is not ended")
)

type Reader interface {
	GetMakeupList(ctx context.Context, actor auth.AuthenticatedUser, contestSlug string) (List, error)
}

type List struct {
	ContestSlug string
	GeneratedAt time.Time
	Items       []Item
}

type Item struct {
	ProblemID       uuid.UUID
	ProblemCode     string
	ProblemSlug     string
	ProblemTitle    string
	Difficulty      string
	Category        string
	LastStatus      string
	SeverityRank    int
	ReasonSummary   string
	SuggestedAction string
	AttemptCount    int
}
