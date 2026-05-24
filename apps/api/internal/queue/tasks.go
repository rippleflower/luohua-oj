package queue

import (
	"time"

	"github.com/google/uuid"
)

const TypeJudgeSubmission = "judge:submission"

const JudgeQueueName = "judge"

type JudgeSubmissionPayload struct {
	SubmissionID          uuid.UUID  `json:"submissionId"`
	ProblemVersionID      *uuid.UUID `json:"problemVersionId,omitempty"`
	ResultSnapshotVersion int        `json:"resultSnapshotVersion"`
}

type EnqueueJudgeSubmissionInput = JudgeSubmissionPayload

type TaskSummary struct {
	ID            string
	Type          string
	SubmissionID  *uuid.UUID
	State         string
	CreatedAt     *time.Time
	NextProcessAt *time.Time
	CompletedAt   *time.Time
	LastErr       string
}

type Summary struct {
	Queue          string
	Paused         bool
	LatencySeconds int
	Pending        int
	Active         int
	Scheduled      int
	Retry          int
	Archived       int
	Completed      int
	ProcessedToday int
	FailedToday    int
	RecentTasks    []TaskSummary
	UpdatedAt      time.Time
}
