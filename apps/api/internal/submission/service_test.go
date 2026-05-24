package submission_test

import (
	"context"
	"io"
	"log/slog"
	"testing"
	"time"

	"github.com/example/oj3/apps/api/internal/auth"
	"github.com/example/oj3/apps/api/internal/queue"
	"github.com/example/oj3/apps/api/internal/submission"
	"github.com/google/uuid"
	"github.com/stretchr/testify/require"
)

func TestCreateSubmissionEnqueuesJudgeJob(t *testing.T) {
	repo := &fakeRepo{}
	store := &fakeSourceStore{object: "submissions/source.cpp.zst"}
	queue := &fakeQueue{}
	service := submission.NewService(repo, store, queue, slog.New(slog.NewJSONHandler(io.Discard, nil)))

	created, err := service.Create(context.Background(), submission.CreateInput{
		UserID:    uuid.New(),
		ProblemID: uuid.New(),
		Language:  "CPP17",
		Source:    "int main(){return 0;}",
	})

	require.NoError(t, err)
	require.Equal(t, submission.StatusPending, created.Status)
	require.Equal(t, created.ID, queue.submissionID)
	require.Equal(t, "submissions/source.cpp.zst", repo.params.SourceObjectKey)
	require.Equal(t, created.ID, repo.refreshedSubmissionID)
}

func TestCreateSubmissionRejectsBlankSource(t *testing.T) {
	service := submission.NewService(&fakeRepo{}, &fakeSourceStore{}, &fakeQueue{}, slog.New(slog.NewJSONHandler(io.Discard, nil)))

	_, err := service.Create(context.Background(), submission.CreateInput{
		UserID:    uuid.New(),
		ProblemID: uuid.New(),
		Language:  "CPP17",
		Source:    "   ",
	})

	require.ErrorContains(t, err, "source is required")
}

type fakeRepo struct {
	params                submission.CreateParams
	refreshedSubmissionID uuid.UUID
	rejudgeTarget         submission.RejudgeTarget
	resetSubmissionID     uuid.UUID
	resetSnapshotVersion  int
}

func (f *fakeRepo) CreateSubmission(ctx context.Context, params submission.CreateParams) (submission.Submission, error) {
	f.params = params
	return submission.Submission{
		ID:              uuid.New(),
		UserID:          params.UserID,
		ProblemID:       params.ProblemID,
		Language:        params.Language,
		SourceObjectKey: params.SourceObjectKey,
		Status:          submission.StatusPending,
		CreatedAt:       time.Now().UTC(),
	}, nil
}

func (f *fakeRepo) GetSubmission(ctx context.Context, submissionID uuid.UUID) (submission.SubmissionDetail, error) {
	return submission.SubmissionDetail{}, nil
}

func (f *fakeRepo) ListSubmissionsByUsername(ctx context.Context, params submission.ListByUsernameParams) (submission.SubmissionPage, error) {
	return submission.SubmissionPage{}, nil
}

func (f *fakeRepo) RefreshSubmissionViews(ctx context.Context, submissionID uuid.UUID) error {
	f.refreshedSubmissionID = submissionID
	return nil
}

func (f *fakeRepo) GetRejudgeTarget(ctx context.Context, submissionID uuid.UUID) (submission.RejudgeTarget, error) {
	return f.rejudgeTarget, nil
}

func (f *fakeRepo) ResetSubmissionForRejudge(ctx context.Context, actor auth.AuthenticatedUser, input submission.RejudgeAdminInput, resultSnapshotVersion int, problemVersionID *uuid.UUID) error {
	f.resetSubmissionID = input.SubmissionID
	f.resetSnapshotVersion = resultSnapshotVersion
	return nil
}

func TestListByUsernameAppliesPaginationDefaults(t *testing.T) {
	repo := &capturingListRepo{}
	service := submission.NewService(repo, &fakeSourceStore{}, &fakeQueue{}, slog.New(slog.NewJSONHandler(io.Discard, nil)))

	_, err := service.ListByUsername(context.Background(), submission.ListByUsernameParams{
		Username: " demo ",
	})

	require.NoError(t, err)
	require.Equal(t, "demo", repo.params.Username)
	require.Equal(t, 1, repo.params.Page)
	require.Equal(t, 20, repo.params.PageSize)
}

type capturingListRepo struct {
	params submission.ListByUsernameParams
}

func (c *capturingListRepo) CreateSubmission(ctx context.Context, params submission.CreateParams) (submission.Submission, error) {
	return submission.Submission{}, nil
}

func (c *capturingListRepo) GetSubmission(ctx context.Context, submissionID uuid.UUID) (submission.SubmissionDetail, error) {
	return submission.SubmissionDetail{}, nil
}

func (c *capturingListRepo) ListSubmissionsByUsername(ctx context.Context, params submission.ListByUsernameParams) (submission.SubmissionPage, error) {
	c.params = params
	return submission.SubmissionPage{}, nil
}

func (c *capturingListRepo) RefreshSubmissionViews(ctx context.Context, submissionID uuid.UUID) error {
	return nil
}

func (c *capturingListRepo) GetRejudgeTarget(ctx context.Context, submissionID uuid.UUID) (submission.RejudgeTarget, error) {
	return submission.RejudgeTarget{}, nil
}

func (c *capturingListRepo) ResetSubmissionForRejudge(ctx context.Context, actor auth.AuthenticatedUser, input submission.RejudgeAdminInput, resultSnapshotVersion int, problemVersionID *uuid.UUID) error {
	return nil
}

type fakeSourceStore struct {
	object string
}

func (f *fakeSourceStore) PutSource(ctx context.Context, submissionID uuid.UUID, language string, source string) (string, error) {
	return f.object, nil
}

type fakeQueue struct {
	submissionID uuid.UUID
	input        queue.EnqueueJudgeSubmissionInput
}

func (f *fakeQueue) EnqueueJudgeSubmission(ctx context.Context, input queue.EnqueueJudgeSubmissionInput) error {
	f.submissionID = input.SubmissionID
	f.input = input
	return nil
}

func (f *fakeQueue) QueueSummary(ctx context.Context) (queue.Summary, error) {
	return queue.Summary{}, nil
}

func TestRejudgeAdminUsesCurrentProblemVersionForStandaloneSubmission(t *testing.T) {
	submissionID := uuid.New()
	currentVersionID := uuid.New()
	repo := &fakeRepo{
		rejudgeTarget: submission.RejudgeTarget{
			SubmissionID:            submissionID,
			CurrentProblemVersionID: &currentVersionID,
		},
	}
	queueClient := &fakeQueue{}
	service := submission.NewService(repo, &fakeSourceStore{}, queueClient, slog.New(slog.NewJSONHandler(io.Discard, nil)))

	result, err := service.RejudgeAdmin(context.Background(), auth.AuthenticatedUser{User: auth.User{ID: uuid.New()}}, submission.RejudgeAdminInput{
		SubmissionID: submissionID,
	})

	require.NoError(t, err)
	require.Equal(t, submissionID, repo.resetSubmissionID)
	require.Zero(t, repo.resetSnapshotVersion)
	require.NotNil(t, queueClient.input.ProblemVersionID)
	require.Equal(t, currentVersionID, *queueClient.input.ProblemVersionID)
	require.Equal(t, queue.JudgeQueueName, result.Queue)
}

func TestRejudgeAdminUsesFrozenSnapshotVersionForContestSubmission(t *testing.T) {
	submissionID := uuid.New()
	contestID := uuid.New()
	snapshotVersionID := uuid.New()
	repo := &fakeRepo{
		rejudgeTarget: submission.RejudgeTarget{
			SubmissionID:             submissionID,
			ContestID:                &contestID,
			ResultSnapshotVersion:    2,
			SnapshotProblemVersionID: &snapshotVersionID,
		},
	}
	queueClient := &fakeQueue{}
	service := submission.NewService(repo, &fakeSourceStore{}, queueClient, slog.New(slog.NewJSONHandler(io.Discard, nil)))

	result, err := service.RejudgeAdmin(context.Background(), auth.AuthenticatedUser{User: auth.User{ID: uuid.New()}}, submission.RejudgeAdminInput{
		SubmissionID: submissionID,
	})

	require.NoError(t, err)
	require.Equal(t, 2, repo.resetSnapshotVersion)
	require.Equal(t, 2, queueClient.input.ResultSnapshotVersion)
	require.NotNil(t, queueClient.input.ProblemVersionID)
	require.Equal(t, snapshotVersionID, *queueClient.input.ProblemVersionID)
	require.Equal(t, 2, result.ResultSnapshotVersion)
}
