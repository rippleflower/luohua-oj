package submission_test

import (
	"context"
	"testing"

	"github.com/example/oj3/apps/api/internal/submission"
	"github.com/google/uuid"
	"github.com/stretchr/testify/require"
)

func TestCreateSubmissionEnqueuesJudgeJob(t *testing.T) {
	repo := &fakeRepo{}
	store := &fakeSourceStore{object: "submissions/source.cpp"}
	queue := &fakeQueue{}
	service := submission.NewService(repo, store, queue)

	created, err := service.Create(context.Background(), submission.CreateInput{
		UserID:    uuid.New(),
		ProblemID: uuid.New(),
		Language:  "CPP17",
		Source:    "int main(){return 0;}",
	})

	require.NoError(t, err)
	require.Equal(t, submission.StatusPending, created.Status)
	require.Equal(t, created.ID, queue.submissionID)
	require.Equal(t, "submissions/source.cpp", repo.params.SourceObject)
}

func TestCreateSubmissionRejectsBlankSource(t *testing.T) {
	service := submission.NewService(&fakeRepo{}, &fakeSourceStore{}, &fakeQueue{})

	_, err := service.Create(context.Background(), submission.CreateInput{
		UserID:    uuid.New(),
		ProblemID: uuid.New(),
		Language:  "CPP17",
		Source:    "   ",
	})

	require.ErrorContains(t, err, "source is required")
}

type fakeRepo struct {
	params submission.CreateParams
}

func (f *fakeRepo) CreateSubmission(ctx context.Context, params submission.CreateParams) (submission.Submission, error) {
	f.params = params
	return submission.Submission{
		ID:           uuid.New(),
		UserID:       params.UserID,
		ProblemID:    params.ProblemID,
		Language:     params.Language,
		SourceObject: params.SourceObject,
		Status:       submission.StatusPending,
	}, nil
}

type fakeSourceStore struct {
	object string
}

func (f *fakeSourceStore) PutSource(ctx context.Context, submissionID uuid.UUID, source string) (string, error) {
	return f.object, nil
}

type fakeQueue struct {
	submissionID uuid.UUID
}

func (f *fakeQueue) EnqueueJudgeSubmission(ctx context.Context, submissionID uuid.UUID) error {
	f.submissionID = submissionID
	return nil
}
