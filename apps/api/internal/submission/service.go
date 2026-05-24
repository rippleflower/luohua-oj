package submission

import (
	"context"
	"errors"
	"log/slog"
	"strings"
	"time"

	"github.com/example/oj3/apps/api/internal/auth"
	"github.com/example/oj3/apps/api/internal/queue"
	"github.com/google/uuid"
)

type Status string

const StatusPending Status = "PENDING"

type Submission struct {
	ID              uuid.UUID
	UserID          uuid.UUID
	ProblemID       uuid.UUID
	Problem         *ProblemSummary
	Language        string
	SourceObjectKey string
	Status          Status
	CreatedAt       time.Time
}

type SubmissionDetail struct {
	Submission
	CompileSummary       CompileSummary
	Results              []ResultSummary
	ArtifactAvailability ArtifactAvailability
}

type ProblemSummary struct {
	ID    uuid.UUID
	Slug  string
	Title string
}

type CompileSummary struct {
	CompileOutput string
	MaxTimeMs     *int32
	MaxMemoryKb   *int32
	JudgedAt      *time.Time
}

type ResultSummary struct {
	TestCaseID    uuid.UUID
	Status        Status
	TimeMs        *int32
	MemoryKb      *int32
	OutputSnippet string
	ErrorSnippet  string
}

type ArtifactSummary struct {
	ArtifactType    string
	ObjectKey       string
	ContentType     string
	ContentEncoding string
	ExpiresAt       *time.Time
}

type ArtifactAvailability struct {
	SourceObjectKey string
	Artifacts       []ArtifactSummary
}

type ListByUsernameParams struct {
	Username string
	Page     int
	PageSize int
}

type SubmissionPage struct {
	Items    []Submission
	Total    int
	Page     int
	PageSize int
}

type CreateInput struct {
	UserID    uuid.UUID
	ProblemID uuid.UUID
	Language  string
	Source    string
}

type CreateParams struct {
	ID              uuid.UUID
	UserID          uuid.UUID
	ProblemID       uuid.UUID
	Language        string
	SourceObjectKey string
}

type RejudgeTarget struct {
	SubmissionID             uuid.UUID
	ContestID                *uuid.UUID
	ResultSnapshotVersion    int
	CurrentProblemVersionID  *uuid.UUID
	SnapshotProblemVersionID *uuid.UUID
}

func (r RejudgeTarget) EffectiveProblemVersionID() *uuid.UUID {
	if r.ContestID != nil && r.ResultSnapshotVersion > 0 && r.SnapshotProblemVersionID != nil {
		return r.SnapshotProblemVersionID
	}
	return r.CurrentProblemVersionID
}

func (r RejudgeTarget) EffectiveResultSnapshotVersion() int {
	if r.ContestID != nil && r.ResultSnapshotVersion > 0 {
		return r.ResultSnapshotVersion
	}
	return 0
}

type RejudgeAdminInput struct {
	SubmissionID uuid.UUID
	Reason       string
	IP           string
}

type RejudgeResult struct {
	SubmissionID          uuid.UUID
	Queue                 string
	ResultSnapshotVersion int
	ProblemVersionID      *uuid.UUID
}

type QueueSummary = queue.Summary

type Repository interface {
	CreateSubmission(ctx context.Context, params CreateParams) (Submission, error)
	GetSubmission(ctx context.Context, submissionID uuid.UUID) (SubmissionDetail, error)
	ListSubmissionsByUsername(ctx context.Context, params ListByUsernameParams) (SubmissionPage, error)
	RefreshSubmissionViews(ctx context.Context, submissionID uuid.UUID) error
	GetRejudgeTarget(ctx context.Context, submissionID uuid.UUID) (RejudgeTarget, error)
	ResetSubmissionForRejudge(ctx context.Context, actor auth.AuthenticatedUser, input RejudgeAdminInput, resultSnapshotVersion int, problemVersionID *uuid.UUID) error
}

type Creator interface {
	Create(ctx context.Context, input CreateInput) (Submission, error)
}

type Reader interface {
	Get(ctx context.Context, submissionID uuid.UUID) (SubmissionDetail, error)
	ListByUsername(ctx context.Context, params ListByUsernameParams) (SubmissionPage, error)
}

type AdminManager interface {
	RejudgeAdmin(ctx context.Context, actor auth.AuthenticatedUser, input RejudgeAdminInput) (RejudgeResult, error)
	QueueSummary(ctx context.Context) (QueueSummary, error)
}

type SourceStore interface {
	PutSource(ctx context.Context, submissionID uuid.UUID, language string, source string) (string, error)
}

type Service struct {
	repo       Repository
	source     SourceStore
	judgeQueue queue.JudgeQueue
	logger     *slog.Logger
}

func NewService(repo Repository, source SourceStore, judgeQueue queue.JudgeQueue, logger *slog.Logger) *Service {
	if logger == nil {
		logger = slog.New(slog.NewTextHandler(ioDiscard{}, nil))
	}
	return &Service{repo: repo, source: source, judgeQueue: judgeQueue, logger: logger}
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
	sourceObjectKey, err := s.source.PutSource(ctx, submissionID, input.Language, input.Source)
	if err != nil {
		return Submission{}, err
	}

	created, err := s.repo.CreateSubmission(ctx, CreateParams{
		ID:              submissionID,
		UserID:          input.UserID,
		ProblemID:       input.ProblemID,
		Language:        input.Language,
		SourceObjectKey: sourceObjectKey,
	})
	if err != nil {
		return Submission{}, err
	}
	if err := s.repo.RefreshSubmissionViews(ctx, created.ID); err != nil {
		return Submission{}, err
	}

	s.logger.InfoContext(ctx,
		"submission created",
		"event", "submission.created",
		"submissionId", created.ID.String(),
		"problemId", created.ProblemID.String(),
		"language", created.Language,
	)

	queueInput := queue.EnqueueJudgeSubmissionInput{SubmissionID: created.ID}
	if err := s.judgeQueue.EnqueueJudgeSubmission(ctx, queueInput); err != nil {
		s.logger.ErrorContext(ctx,
			"judge queue enqueue failed",
			"event", "submission.enqueue.failed",
			"submissionId", created.ID.String(),
			"error", err,
		)
		return Submission{}, err
	}

	s.logger.InfoContext(ctx,
		"judge queue enqueued",
		"event", "submission.enqueue.succeeded",
		"submissionId", created.ID.String(),
	)

	return created, nil
}

func (s *Service) Get(ctx context.Context, submissionID uuid.UUID) (SubmissionDetail, error) {
	if submissionID == uuid.Nil {
		return SubmissionDetail{}, errors.New("submission id is required")
	}

	return s.repo.GetSubmission(ctx, submissionID)
}

const (
	defaultListPage     = 1
	defaultListPageSize = 20
	maxListPageSize     = 100
)

func (s *Service) ListByUsername(ctx context.Context, params ListByUsernameParams) (SubmissionPage, error) {
	if strings.TrimSpace(params.Username) == "" {
		return SubmissionPage{}, errors.New("username is required")
	}

	if params.Page <= 0 {
		params.Page = defaultListPage
	}
	if params.PageSize <= 0 {
		params.PageSize = defaultListPageSize
	}
	if params.PageSize > maxListPageSize {
		params.PageSize = maxListPageSize
	}

	params.Username = strings.TrimSpace(params.Username)
	return s.repo.ListSubmissionsByUsername(ctx, params)
}

func (s *Service) RejudgeAdmin(ctx context.Context, actor auth.AuthenticatedUser, input RejudgeAdminInput) (RejudgeResult, error) {
	if input.SubmissionID == uuid.Nil {
		return RejudgeResult{}, errors.New("submission id is required")
	}

	target, err := s.repo.GetRejudgeTarget(ctx, input.SubmissionID)
	if err != nil {
		return RejudgeResult{}, err
	}

	queueInput := queue.EnqueueJudgeSubmissionInput{
		SubmissionID:          target.SubmissionID,
		ProblemVersionID:      target.EffectiveProblemVersionID(),
		ResultSnapshotVersion: target.EffectiveResultSnapshotVersion(),
	}
	if err := s.repo.ResetSubmissionForRejudge(ctx, actor, input, queueInput.ResultSnapshotVersion, queueInput.ProblemVersionID); err != nil {
		return RejudgeResult{}, err
	}
	if err := s.judgeQueue.EnqueueJudgeSubmission(ctx, queueInput); err != nil {
		return RejudgeResult{}, err
	}

	s.logger.InfoContext(ctx,
		"submission rejudge enqueued",
		"event", "submission.rejudge.enqueued",
		"submissionId", input.SubmissionID.String(),
		"actorUserId", actor.ID.String(),
		"resultSnapshotVersion", queueInput.ResultSnapshotVersion,
	)

	return RejudgeResult{
		SubmissionID:          target.SubmissionID,
		Queue:                 queue.JudgeQueueName,
		ResultSnapshotVersion: queueInput.ResultSnapshotVersion,
		ProblemVersionID:      queueInput.ProblemVersionID,
	}, nil
}

func (s *Service) QueueSummary(ctx context.Context) (QueueSummary, error) {
	return s.judgeQueue.QueueSummary(ctx)
}

type ioDiscard struct{}

func (ioDiscard) Write(p []byte) (int, error) {
	return len(p), nil
}
