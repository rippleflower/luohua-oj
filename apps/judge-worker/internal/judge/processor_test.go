package judge_test

import (
	"context"
	"encoding/json"
	"errors"
	"log/slog"
	"strings"
	"testing"

	"github.com/example/oj3/apps/judge-worker/internal/judge"
	"github.com/example/oj3/apps/judge-worker/internal/queue"
	"github.com/google/uuid"
	"github.com/hibiken/asynq"
	"github.com/stretchr/testify/require"
)

var errExecutorFailed = errors.New("executor failed")

func TestHandleJudgeSubmissionAcceptsAllPassingTests(t *testing.T) {
	submissionID := uuid.New()
	problemID := uuid.New()
	repo := &fakeRepo{
		submission: judge.Submission{
			ID:            submissionID,
			ProblemID:     problemID,
			Language:      "CPP17",
			SourceObject:  "tmp/source.cpp",
			TimeLimitMs:   1000,
			MemoryLimitKB: 262144,
		},
		testCases: []judge.TestCase{
			{ID: uuid.New(), InputObject: "in-1", OutputObject: "out-1"},
			{ID: uuid.New(), InputObject: "in-2", OutputObject: "out-2"},
		},
	}
	executor := &fakeExecutor{outcome: judge.JudgeOutcome{
		Status: judge.StatusAccepted,
		Results: []judge.RunResult{
			{Status: judge.StatusAccepted, Output: "out-1", TimeMs: 12, MemoryKB: 32},
			{Status: judge.StatusAccepted, Output: "out-2", TimeMs: 18, MemoryKB: 64},
		},
	}}
	processor := judge.NewProcessor(repo, executor, nil)

	err := processor.HandleJudgeSubmission(context.Background(), taskForSubmission(t, submissionID))

	require.NoError(t, err)
	require.Equal(t, []judge.Status{judge.StatusRunning, judge.StatusAccepted}, repo.statuses)
	require.Len(t, repo.savedResults, 2)
	require.Equal(t, 1, executor.calls)
	require.NotNil(t, repo.maxTimeMs)
	require.NotNil(t, repo.maxMemoryKB)
	require.Equal(t, int32(18), *repo.maxTimeMs)
	require.Equal(t, int32(64), *repo.maxMemoryKB)
}

func TestHandleJudgeSubmissionStopsAfterWrongAnswer(t *testing.T) {
	submissionID := uuid.New()
	problemID := uuid.New()
	repo := &fakeRepo{
		submission: judge.Submission{ID: submissionID, ProblemID: problemID, TimeLimitMs: 1000, MemoryLimitKB: 262144},
		testCases: []judge.TestCase{
			{ID: uuid.New(), InputObject: "in-1", OutputObject: "out-1"},
			{ID: uuid.New(), InputObject: "in-2", OutputObject: "out-2"},
		},
	}
	executor := &fakeExecutor{outcome: judge.JudgeOutcome{
		Status:  judge.StatusWrongAnswer,
		Results: []judge.RunResult{{Status: judge.StatusWrongAnswer, Output: "bad"}},
	}}
	processor := judge.NewProcessor(repo, executor, nil)

	err := processor.HandleJudgeSubmission(context.Background(), taskForSubmission(t, submissionID))

	require.NoError(t, err)
	require.Equal(t, []judge.Status{judge.StatusRunning, judge.StatusWrongAnswer}, repo.statuses)
	require.Len(t, repo.savedResults, 1)
	require.Equal(t, 1, executor.calls)
}

func TestHandleJudgeSubmissionMarksSystemErrorWithoutTestCases(t *testing.T) {
	submissionID := uuid.New()
	repo := &fakeRepo{
		submission: judge.Submission{ID: submissionID, ProblemID: uuid.New(), TimeLimitMs: 1000, MemoryLimitKB: 262144},
	}
	processor := judge.NewProcessor(repo, &fakeExecutor{}, nil)

	err := processor.HandleJudgeSubmission(context.Background(), taskForSubmission(t, submissionID))

	require.ErrorContains(t, err, "problem has no test cases")
	require.Equal(t, []judge.Status{judge.StatusRunning, judge.StatusSystemError}, repo.statuses)
	require.Equal(t, "problem has no test cases", repo.compileOutput)
}

func TestHandleJudgeSubmissionSavesResultWhenExecutorFails(t *testing.T) {
	submissionID := uuid.New()
	testCaseID := uuid.New()
	repo := &fakeRepo{
		submission: judge.Submission{ID: submissionID, ProblemID: uuid.New(), TimeLimitMs: 1000, MemoryLimitKB: 262144},
		testCases:  []judge.TestCase{{ID: testCaseID}},
	}
	processor := judge.NewProcessor(repo, &fakeExecutor{err: errExecutorFailed}, nil)

	err := processor.HandleJudgeSubmission(context.Background(), taskForSubmission(t, submissionID))

	require.ErrorIs(t, err, errExecutorFailed)
	require.Equal(t, []judge.Status{judge.StatusRunning, judge.StatusSystemError}, repo.statuses)
	require.Empty(t, repo.savedResults)
	require.Equal(t, "judge submission: executor failed", repo.compileOutput)
}

func TestHandleJudgeSubmissionRejectsMissingSubmissionID(t *testing.T) {
	repo := &fakeRepo{}
	payload, err := json.Marshal(queue.JudgeSubmissionPayload{})
	require.NoError(t, err)
	processor := judge.NewProcessor(repo, &fakeExecutor{}, nil)

	err = processor.HandleJudgeSubmission(context.Background(), asynq.NewTask(queue.TypeJudgeSubmission, payload))

	require.ErrorContains(t, err, "submissionId is required")
	require.Empty(t, repo.statuses)
}

func TestHandleJudgeSubmissionLogsCompletion(t *testing.T) {
	submissionID := uuid.New()
	repo := &fakeRepo{
		submission: judge.Submission{ID: submissionID, ProblemID: uuid.New(), TimeLimitMs: 1000, MemoryLimitKB: 262144},
		testCases:  []judge.TestCase{{ID: uuid.New()}},
	}
	executor := &fakeExecutor{outcome: judge.JudgeOutcome{
		Status:  judge.StatusAccepted,
		Results: []judge.RunResult{{Status: judge.StatusAccepted}},
	}}
	var buffer strings.Builder
	logger := slog.New(slog.NewJSONHandler(&buffer, nil))
	processor := judge.NewProcessor(repo, executor, logger)

	err := processor.HandleJudgeSubmission(context.Background(), taskForSubmission(t, submissionID))

	require.NoError(t, err)
	require.Contains(t, buffer.String(), `"event":"judge.task.completed"`)
	require.Contains(t, buffer.String(), submissionID.String())
}

func taskForSubmission(t *testing.T, submissionID uuid.UUID) *asynq.Task {
	t.Helper()
	payload, err := json.Marshal(queue.JudgeSubmissionPayload{SubmissionID: submissionID})
	require.NoError(t, err)
	return asynq.NewTask(queue.TypeJudgeSubmission, payload)
}

type savedResult struct {
	submissionID uuid.UUID
	testCaseID   uuid.UUID
	result       judge.RunResult
}

type fakeRepo struct {
	submission    judge.Submission
	testCases     []judge.TestCase
	statuses      []judge.Status
	compileOutput string
	maxTimeMs     *int32
	maxMemoryKB   *int32
	savedResults  []savedResult
}

func (f *fakeRepo) GetSubmission(ctx context.Context, submissionID uuid.UUID) (judge.Submission, error) {
	return f.submission, nil
}

func (f *fakeRepo) ListTestCases(ctx context.Context, problemID uuid.UUID) ([]judge.TestCase, error) {
	return f.testCases, nil
}

func (f *fakeRepo) SaveResult(ctx context.Context, submissionID uuid.UUID, testCaseID uuid.UUID, result judge.RunResult) error {
	f.savedResults = append(f.savedResults, savedResult{
		submissionID: submissionID,
		testCaseID:   testCaseID,
		result:       result,
	})
	return nil
}

func (f *fakeRepo) UpdateStatus(ctx context.Context, submissionID uuid.UUID, status judge.Status, compileOutput string, maxTimeMs *int32, maxMemoryKB *int32) error {
	f.statuses = append(f.statuses, status)
	f.compileOutput = compileOutput
	f.maxTimeMs = maxTimeMs
	f.maxMemoryKB = maxMemoryKB
	return nil
}

type fakeExecutor struct {
	outcome judge.JudgeOutcome
	calls   int
	err     error
}

func (f *fakeExecutor) Judge(ctx context.Context, submission judge.Submission, testCases []judge.TestCase) (judge.JudgeOutcome, error) {
	if f.err != nil {
		return judge.JudgeOutcome{}, f.err
	}
	f.calls++
	return f.outcome, nil
}
