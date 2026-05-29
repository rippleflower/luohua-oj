package contest_makeup_test

import (
	"context"
	"testing"

	"github.com/example/oj3/apps/api/internal/auth"
	"github.com/example/oj3/apps/api/internal/contest_makeup"
	"github.com/google/uuid"
	"github.com/stretchr/testify/require"
)

func TestGetMakeupListRequiresEndedContest(t *testing.T) {
	repo := &fakeRepository{
		meta: contest_makeup.ContestMeta{
			ID:     uuid.New(),
			Slug:   "spring-open",
			Status: "RUNNING",
		},
	}
	service := contest_makeup.NewService(repo)

	_, err := service.GetMakeupList(context.Background(), auth.AuthenticatedUser{
		User: auth.User{ID: uuid.New()},
	}, "spring-open")

	require.ErrorIs(t, err, contest_makeup.ErrContestNotEnded)
}

func TestGetMakeupListSortsByCategoryDifficultySeverityThenCode(t *testing.T) {
	contestID := uuid.New()
	problemA := uuid.New()
	problemB := uuid.New()
	problemC := uuid.New()
	problemD := uuid.New()
	repo := &fakeRepository{
		meta: contest_makeup.ContestMeta{
			ID:     contestID,
			Slug:   "april-grand-prix",
			Status: "ENDED",
		},
		problems: []contest_makeup.ContestProblem{
			{ProblemID: problemA, ProblemCode: "A", ProblemTitle: "A", Difficulty: "EASY"},
			{ProblemID: problemB, ProblemCode: "B", ProblemTitle: "B", Difficulty: "MEDIUM"},
			{ProblemID: problemC, ProblemCode: "C", ProblemTitle: "C", Difficulty: "HARD"},
			{ProblemID: problemD, ProblemCode: "D", ProblemTitle: "D", Difficulty: "MEDIUM"},
		},
		attempts: map[uuid.UUID]contest_makeup.ProblemAttempt{
			problemB: {ProblemID: problemB, HasAccepted: false, LatestStatus: "WRONG_ANSWER", AttemptCount: 2},
			problemC: {ProblemID: problemC, HasAccepted: false, LatestStatus: "TIME_LIMIT_EXCEEDED", AttemptCount: 1},
			problemD: {ProblemID: problemD, HasAccepted: false, LatestStatus: "RUNTIME_ERROR", AttemptCount: 3},
		},
	}
	service := contest_makeup.NewService(repo)

	list, err := service.GetMakeupList(context.Background(), auth.AuthenticatedUser{
		User: auth.User{ID: uuid.New()},
	}, "april-grand-prix")

	require.NoError(t, err)
	require.Len(t, list.Items, 4)
	require.Equal(t, "ATTEMPTED_UNSOLVED", list.Items[0].Category)
	require.Equal(t, "D", list.Items[0].ProblemCode)
	require.Equal(t, 0, list.Items[0].SeverityRank)
	require.Equal(t, 3, list.Items[0].AttemptCount)
	require.Equal(t, "ATTEMPTED_UNSOLVED", list.Items[1].Category)
	require.Equal(t, "B", list.Items[1].ProblemCode)
	require.Equal(t, "ATTEMPTED_UNSOLVED", list.Items[2].Category)
	require.Equal(t, "C", list.Items[2].ProblemCode)
	require.Equal(t, "UNATTEMPTED_RECOMMENDED", list.Items[3].Category)
	require.Equal(t, "A", list.Items[3].ProblemCode)
	require.Equal(t, 99, list.Items[3].SeverityRank)
	require.Equal(t, 0, list.Items[3].AttemptCount)
}

type fakeRepository struct {
	meta     contest_makeup.ContestMeta
	problems []contest_makeup.ContestProblem
	attempts map[uuid.UUID]contest_makeup.ProblemAttempt
}

func (f *fakeRepository) GetContestMeta(ctx context.Context, slug string) (contest_makeup.ContestMeta, error) {
	return f.meta, nil
}

func (f *fakeRepository) ListContestProblems(ctx context.Context, contestID uuid.UUID) ([]contest_makeup.ContestProblem, error) {
	return f.problems, nil
}

func (f *fakeRepository) ListUserProblemAttempts(ctx context.Context, userID uuid.UUID, contestID uuid.UUID) (map[uuid.UUID]contest_makeup.ProblemAttempt, error) {
	return f.attempts, nil
}
