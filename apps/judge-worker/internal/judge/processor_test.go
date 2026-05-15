package judge_test

import (
	"context"
	"encoding/json"
	"errors"
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
			ID:           submissionID,
			ProblemID:    problemID,
			Language:     "CPP17",
			SourceObject: "tmp/source.cpp",
		},
		testCases: []judge.TestCase{
			{ID: uuid.New(), InputObject: "in-1", OutputObject: "out-1"},
			{ID: uuid.New(), InputObject: "in-2", OutputObject: "out-2"},
		},
	}
	executor := &fakeExecutor{results: []judge.RunResult{
		{Status: judge.StatusAccepted, Output: "out-1"},
		{Status: judge.StatusAccepted, Output: "out-2"},
	}}
	processor := judge.NewProcessor(repo, executor)

	err := processor.HandleJudgeSubmission(context.Background(), taskForSubmission(t, submissionID))

	require.NoError(t, err)
	require.Equal(t, []judge.Status{judge.StatusRunning, judge.StatusAccepted}, repo.statuses)
	require.Len(t, repo.savedResults, 2)
	require.Equal(t, 2, executor.calls)
}

func TestHandleJudgeSubmissionStopsAfterWrongAnswer(t *testing.T) {
	submissionID := uuid.New()
	problemID := uuid.New()
	repo := &fakeRepo{
		submission: judge.Submission{ID: submissionID, ProblemID: problemID},
		testCases: []judge.TestCase{
			{ID: uuid.New(), InputObject: "in-1", OutputObject: "out-1"},
			{ID: uuid.New(), InputObject: "in-2", OutputObject: "out-2"},
		},
	}
	executor := &fakeExecutor{results: []judge.RunResult{
		{Status: judge.StatusWrongAnswer, Output: "bad"},
	}}
	processor := judge.NewProcessor(repo, executor)

	err := processor.HandleJudgeSubmission(context.Background(), taskForSubmission(t, submissionID))

	require.NoError(t, err)
	require.Equal(t, []judge.Status{judge.StatusRunning, judge.StatusWrongAnswer}, repo.statuses)
	require.Len(t, repo.savedResults, 1)
	require.Equal(t, 1, executor.calls)
}

func TestHandleJudgeSubmissionMarksSystemErrorWithoutTestCases(t *testing.T) {
	submissionID := uuid.New()
	repo := &fakeRepo{
		submission: judge.Submission{ID: submissionID, ProblemID: uuid.New()},
	}
	processor := judge.NewProcessor(repo, &fakeExecutor{})

	err := processor.HandleJudgeSubmission(context.Background(), taskForSubmission(t, submissionID))

	require.ErrorContains(t, err, "problem has no test cases")
	require.Equal(t, []judge.Status{judge.StatusRunning, judge.StatusSystemError}, repo.statuses)
	require.Equal(t, "problem has no test cases", repo.compileOutput)
}

func TestHandleJudgeSubmissionSavesResultWhenExecutorFails(t *testing.T) {
	submissionID := uuid.New()
	testCaseID := uuid.New()
	repo := &fakeRepo{
		submission: judge.Submission{ID: submissionID, ProblemID: uuid.New()},
		testCases:  []judge.TestCase{{ID: testCaseID}},
	}
	processor := judge.NewProcessor(repo, &fakeExecutor{err: errExecutorFailed})

	err := processor.HandleJudgeSubmission(context.Background(), taskForSubmission(t, submissionID))

	require.ErrorIs(t, err, errExecutorFailed)
	require.Equal(t, []judge.Status{judge.StatusRunning, judge.StatusSystemError}, repo.statuses)
	require.Len(t, repo.savedResults, 1)
	require.Equal(t, judge.StatusSystemError, repo.savedResults[0].result.Status)
	require.Equal(t, "executor failed", repo.savedResults[0].result.Error)
}

func TestHandleJudgeSubmissionRejectsMissingSubmissionID(t *testing.T) {
	repo := &fakeRepo{}
	payload, err := json.Marshal(queue.JudgeSubmissionPayload{})
	require.NoError(t, err)
	processor := judge.NewProcessor(repo, &fakeExecutor{})

	err = processor.HandleJudgeSubmission(context.Background(), asynq.NewTask(queue.TypeJudgeSubmission, payload))

	require.ErrorContains(t, err, "submissionId is required")
	require.Empty(t, repo.statuses)
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

func (f *fakeRepo) UpdateStatus(ctx context.Context, submissionID uuid.UUID, status judge.Status, compileOutput string) error {
	f.statuses = append(f.statuses, status)
	f.compileOutput = compileOutput
	return nil
}

type fakeExecutor struct {
	results []judge.RunResult
	calls   int
	err     error
}

func (f *fakeExecutor) Run(ctx context.Context, submission judge.Submission, testCase judge.TestCase) (judge.RunResult, error) {
	if f.err != nil {
		return judge.RunResult{}, f.err
	}
	result := f.results[f.calls]
	f.calls++
	return result, nil
}
