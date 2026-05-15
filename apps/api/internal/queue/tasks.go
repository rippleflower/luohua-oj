package queue

import "github.com/google/uuid"

const TypeJudgeSubmission = "judge:submission"

type JudgeSubmissionPayload struct {
	SubmissionID uuid.UUID `json:"submissionId"`
}
