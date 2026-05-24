package judge

import (
	"context"
	"encoding/json"
	"fmt"
	"log/slog"
	"time"

	"github.com/example/oj3/apps/judge-worker/internal/queue"
	"github.com/google/uuid"
	"github.com/hibiken/asynq"
)

type Status string

const (
	StatusAccepted            Status = "ACCEPTED"
	StatusRunning             Status = "RUNNING"
	StatusWrongAnswer         Status = "WRONG_ANSWER"
	StatusTimeLimitExceeded   Status = "TIME_LIMIT_EXCEEDED"
	StatusMemoryLimitExceeded Status = "MEMORY_LIMIT_EXCEEDED"
	StatusRuntimeError        Status = "RUNTIME_ERROR"
	StatusCompileError        Status = "COMPILE_ERROR"
	StatusSystemError         Status = "SYSTEM_ERROR"
)

type Submission struct {
	ID            uuid.UUID
	ProblemID     uuid.UUID
	Language      string
	SourceObject  string
	TimeLimitMs   int32
	MemoryLimitKB int32
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
	UpdateStatus(ctx context.Context, submissionID uuid.UUID, status Status, compileOutput string, maxTimeMs *int32, maxMemoryKB *int32) error
}

type Executor interface {
	Judge(ctx context.Context, submission Submission, testCases []TestCase) (JudgeOutcome, error)
}

type Processor struct {
	repo     Repository
	executor Executor
	logger   *slog.Logger
}

func NewProcessor(repo Repository, executor Executor, logger *slog.Logger) *Processor {
	if logger == nil {
		logger = slog.Default()
	}
	return &Processor{repo: repo, executor: executor, logger: logger}
}

func (p *Processor) Register(mux *asynq.ServeMux) {
	mux.HandleFunc(queue.TypeJudgeSubmission, p.HandleJudgeSubmission)
}

func (p *Processor) HandleJudgeSubmission(ctx context.Context, task *asynq.Task) error {
	startedAt := time.Now()
	taskID, ok := asynq.GetTaskID(ctx)
	if !ok && task.ResultWriter() != nil {
		taskID = task.ResultWriter().TaskID()
	}

	var payload queue.JudgeSubmissionPayload
	if err := json.Unmarshal(task.Payload(), &payload); err != nil {
		p.logger.ErrorContext(ctx, "judge payload decode failed", "event", "judge.task.invalid_payload", "taskId", taskID, "error", err)
		return fmt.Errorf("decode judge payload: %w", err)
	}
	if payload.SubmissionID == uuid.Nil {
		p.logger.ErrorContext(ctx, "judge task missing submissionId", "event", "judge.task.missing_submission_id", "taskId", taskID)
		return fmt.Errorf("submissionId is required")
	}

	p.logger.InfoContext(ctx,
		"judge task received",
		"event", "judge.task.received",
		"taskId", taskID,
		"submissionId", payload.SubmissionID.String(),
		"resultSnapshotVersion", payload.ResultSnapshotVersion,
		"problemVersionId", nullableUUIDString(payload.ProblemVersionID),
	)

	if err := p.repo.UpdateStatus(ctx, payload.SubmissionID, StatusRunning, "", nil, nil); err != nil {
		p.logger.ErrorContext(ctx, "mark running failed", "event", "judge.task.mark_running_failed", "taskId", taskID, "submissionId", payload.SubmissionID.String(), "error", err)
		return fmt.Errorf("mark running: %w", err)
	}

	p.logger.InfoContext(ctx,
		"judge task started",
		"event", "judge.task.started",
		"taskId", taskID,
		"submissionId", payload.SubmissionID.String(),
	)

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

	outcome, err := p.executor.Judge(ctx, submission, testCases)
	if err != nil {
		return p.failSubmission(ctx, payload.SubmissionID, fmt.Errorf("judge submission: %w", err))
	}

	if len(outcome.Results) > len(testCases) {
		return p.failSubmission(ctx, payload.SubmissionID, fmt.Errorf("judge produced %d results for %d test cases", len(outcome.Results), len(testCases)))
	}

	for index, result := range outcome.Results {
		testCase := testCases[index]
		if err := p.repo.SaveResult(ctx, submission.ID, testCase.ID, result); err != nil {
			return p.failSubmission(ctx, payload.SubmissionID, fmt.Errorf("save result: %w", err))
		}
	}

	maxTimeMs, maxMemoryKB := summarizeResults(outcome.Results)
	if err := p.repo.UpdateStatus(ctx, submission.ID, outcome.Status, outcome.CompileOutput, maxTimeMs, maxMemoryKB); err != nil {
		p.logger.ErrorContext(ctx, "judge task completion failed", "event", "judge.task.complete_failed", "taskId", taskID, "submissionId", submission.ID.String(), "error", err)
		return err
	}

	p.logger.InfoContext(ctx,
		"judge task completed",
		"event", "judge.task.completed",
		"taskId", taskID,
		"submissionId", submission.ID.String(),
		"status", string(outcome.Status),
		"durationMs", time.Since(startedAt).Milliseconds(),
	)

	return nil
}

func nullableUUIDString(value *uuid.UUID) string {
	if value == nil {
		return ""
	}
	return value.String()
}

func (p *Processor) failSubmission(ctx context.Context, submissionID uuid.UUID, cause error) error {
	p.logger.ErrorContext(ctx,
		"judge task failed",
		"event", "judge.task.failed",
		"submissionId", submissionID.String(),
		"error", cause,
	)
	if err := p.repo.UpdateStatus(ctx, submissionID, StatusSystemError, cause.Error(), nil, nil); err != nil {
		return fmt.Errorf("%w; additionally failed to mark system error: %v", cause, err)
	}
	return cause
}

func summarizeResults(results []RunResult) (*int32, *int32) {
	var maxTimeMs int32
	var maxMemoryKB int32
	var hasTime bool
	var hasMemory bool

	for _, result := range results {
		if result.TimeMs > 0 && (!hasTime || result.TimeMs > maxTimeMs) {
			maxTimeMs = result.TimeMs
			hasTime = true
		}
		if result.MemoryKB > 0 && (!hasMemory || result.MemoryKB > maxMemoryKB) {
			maxMemoryKB = result.MemoryKB
			hasMemory = true
		}
	}

	var timePtr *int32
	var memoryPtr *int32
	if hasTime {
		timePtr = &maxTimeMs
	}
	if hasMemory {
		memoryPtr = &maxMemoryKB
	}
	return timePtr, memoryPtr
}
