package queue

import (
	"context"
	"encoding/json"

	"github.com/google/uuid"
	"github.com/hibiken/asynq"
)

type JudgeQueue interface {
	EnqueueJudgeSubmission(ctx context.Context, submissionID uuid.UUID) error
}

type Client struct {
	asynq *asynq.Client
}

func NewClient(redisAddr string) *Client {
	return &Client{
		asynq: asynq.NewClient(asynq.RedisClientOpt{Addr: redisAddr}),
	}
}

func (c *Client) Close() error {
	return c.asynq.Close()
}

func (c *Client) EnqueueJudgeSubmission(ctx context.Context, submissionID uuid.UUID) error {
	payload, err := json.Marshal(JudgeSubmissionPayload{SubmissionID: submissionID})
	if err != nil {
		return err
	}

	_, err = c.asynq.EnqueueContext(
		ctx,
		asynq.NewTask(TypeJudgeSubmission, payload),
		asynq.Queue("judge"),
		asynq.MaxRetry(3),
	)
	return err
}
