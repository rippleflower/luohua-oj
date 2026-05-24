package queue

import "github.com/google/uuid"

const TypeJudgeSubmission = "judge:submission"

type JudgeSubmissionPayload struct {
	SubmissionID          uuid.UUID  `json:"submissionId"`
	ProblemVersionID      *uuid.UUID `json:"problemVersionId,omitempty"`
	ResultSnapshotVersion int        `json:"resultSnapshotVersion"`
}
