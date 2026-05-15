package submission

import (
	"context"
	"errors"
	"strings"

	"github.com/example/oj3/apps/api/internal/queue"
	"github.com/google/uuid"
)

type Status string

const StatusPending Status = "PENDING"

type Submission struct {
	ID           uuid.UUID
	UserID       uuid.UUID
	ProblemID    uuid.UUID
	Language     string
	SourceObject string
	Status       Status
}

type CreateInput struct {
	UserID    uuid.UUID
	ProblemID uuid.UUID
	Language  string
	Source    string
}

type CreateParams struct {
	ID           uuid.UUID
	UserID       uuid.UUID
	ProblemID    uuid.UUID
	Language     string
	SourceObject string
}

type Repository interface {
	CreateSubmission(ctx context.Context, params CreateParams) (Submission, error)
}

type Creator interface {
	Create(ctx context.Context, input CreateInput) (Submission, error)
}

type SourceStore interface {
	PutSource(ctx context.Context, submissionID uuid.UUID, source string) (string, error)
}

type Service struct {
	repo       Repository
	source     SourceStore
	judgeQueue queue.JudgeQueue
}

func NewService(repo Repository, source SourceStore, judgeQueue queue.JudgeQueue) *Service {
	return &Service{repo: repo, source: source, judgeQueue: judgeQueue}
}

func (s *Service) Create(ctx context.Context, input CreateInput) (Submission, error) {
	if input.UserID == uuid.Nil {
		return Submission{}, errors.New("user id is required")
	}
	if input.ProblemID == uuid.Nil {
		return Submission{}, errors.New("problem id is required")
	}
	if strings.TrimSpace(input.Source) == "" {
		return Submission{}, errors.New("source is required")
	}

	submissionID := uuid.New()
	sourceObject, err := s.source.PutSource(ctx, submissionID, input.Source)
	if err != nil {
		return Submission{}, err
	}

	created, err := s.repo.CreateSubmission(ctx, CreateParams{
		ID:           submissionID,
		UserID:       input.UserID,
		ProblemID:    input.ProblemID,
		Language:     input.Language,
		SourceObject: sourceObject,
	})
	if err != nil {
		return Submission{}, err
	}

	if err := s.judgeQueue.EnqueueJudgeSubmission(ctx, created.ID); err != nil {
		return Submission{}, err
	}

	return created, nil
}
