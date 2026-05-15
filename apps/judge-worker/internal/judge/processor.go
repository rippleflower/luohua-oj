package judge

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/example/oj3/apps/judge-worker/internal/queue"
	"github.com/google/uuid"
	"github.com/hibiken/asynq"
)

type Status string

const (
	StatusAccepted    Status = "ACCEPTED"
	StatusRunning     Status = "RUNNING"
	StatusWrongAnswer Status = "WRONG_ANSWER"
	StatusSystemError Status = "SYSTEM_ERROR"
)

type Submission struct {
	ID           uuid.UUID
	ProblemID    uuid.UUID
	Language     string
	SourceObject string
}

type TestCase struct {
	ID           uuid.UUID
	InputObject  string
	OutputObject string
}

type RunResult struct {
	Status   Status
	Output   string
	Error    string
	TimeMs   int32
	MemoryKB int32
}

type Repository interface {
	GetSubmission(ctx context.Context, submissionID uuid.UUID) (Submission, error)
	ListTestCases(ctx context.Context, problemID uuid.UUID) ([]TestCase, error)
	SaveResult(ctx context.Context, submissionID uuid.UUID, testCaseID uuid.UUID, result RunResult) error
	UpdateStatus(ctx context.Context, submissionID uuid.UUID, status Status, compileOutput string) error
}

type Executor interface {
	Run(ctx context.Context, submission Submission, testCase TestCase) (RunResult, error)
}

type Processor struct {
	repo     Repository
	executor Executor
}

func NewProcessor(repo Repository, executor Executor) *Processor {
	return &Processor{repo: repo, executor: executor}
}

func (p *Processor) Register(mux *asynq.ServeMux) {
	mux.HandleFunc(queue.TypeJudgeSubmission, p.HandleJudgeSubmission)
}

func (p *Processor) HandleJudgeSubmission(ctx context.Context, task *asynq.Task) error {
	var payload queue.JudgeSubmissionPayload
	if err := json.Unmarshal(task.Payload(), &payload); err != nil {
		return fmt.Errorf("decode judge payload: %w", err)
	}
	if payload.SubmissionID == uuid.Nil {
		return fmt.Errorf("submissionId is required")
	}

	if err := p.repo.UpdateStatus(ctx, payload.SubmissionID, StatusRunning, ""); err != nil {
		return fmt.Errorf("mark running: %w", err)
	}

	submission, err := p.repo.GetSubmission(ctx, payload.SubmissionID)
	if err != nil {
		return p.failSubmission(ctx, payload.SubmissionID, fmt.Errorf("load submission: %w", err))
	}

	testCases, err := p.repo.ListTestCases(ctx, submission.ProblemID)
	if err != nil {
		return p.failSubmission(ctx, payload.SubmissionID, fmt.Errorf("load test cases: %w", err))
	}
	if len(testCases) == 0 {
		return p.failSubmission(ctx, payload.SubmissionID, fmt.Errorf("problem has no test cases"))
	}

	finalStatus := StatusAccepted
	for _, testCase := range testCases {
		result, err := p.executor.Run(ctx, submission, testCase)
		if err != nil {
			_ = p.repo.SaveResult(ctx, submission.ID, testCase.ID, RunResult{
				Status: StatusSystemError,
				Error:  err.Error(),
			})
			return p.failSubmission(ctx, payload.SubmissionID, fmt.Errorf("run test case: %w", err))
		}
		if err := p.repo.SaveResult(ctx, submission.ID, testCase.ID, result); err != nil {
			return p.failSubmission(ctx, payload.SubmissionID, fmt.Errorf("save result: %w", err))
		}
		if result.Status != StatusAccepted {
			finalStatus = result.Status
			break
		}
	}

	return p.repo.UpdateStatus(ctx, submission.ID, finalStatus, "")
}

func (p *Processor) failSubmission(ctx context.Context, submissionID uuid.UUID, cause error) error {
	if err := p.repo.UpdateStatus(ctx, submissionID, StatusSystemError, cause.Error()); err != nil {
		return fmt.Errorf("%w; additionally failed to mark system error: %v", cause, err)
	}
	return cause
}
