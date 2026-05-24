package queue

import (
	"context"
	"encoding/json"
	"errors"
	"time"

	"github.com/google/uuid"
	"github.com/hibiken/asynq"
)

type JudgeQueue interface {
	EnqueueJudgeSubmission(ctx context.Context, input EnqueueJudgeSubmissionInput) error
	QueueSummary(ctx context.Context) (Summary, error)
}

type Client struct {
	asynq     *asynq.Client
	inspector *asynq.Inspector
}

func NewClient(redisAddr string) *Client {
	redisOpt := asynq.RedisClientOpt{Addr: redisAddr}
	return &Client{
		asynq:     asynq.NewClient(redisOpt),
		inspector: asynq.NewInspector(redisOpt),
	}
}

func (c *Client) Close() error {
	if err := c.asynq.Close(); err != nil {
		return err
	}
	return c.inspector.Close()
}

func (c *Client) EnqueueJudgeSubmission(ctx context.Context, input EnqueueJudgeSubmissionInput) error {
	if input.SubmissionID == uuid.Nil {
		return errors.New("submission id is required")
	}
	payload, err := json.Marshal(JudgeSubmissionPayload{
		SubmissionID:          input.SubmissionID,
		ProblemVersionID:      input.ProblemVersionID,
		ResultSnapshotVersion: input.ResultSnapshotVersion,
	})
	if err != nil {
		return err
	}

	_, err = c.asynq.EnqueueContext(
		ctx,
		asynq.NewTask(TypeJudgeSubmission, payload),
		asynq.Queue(JudgeQueueName),
		asynq.MaxRetry(3),
	)
	return err
}

func (c *Client) QueueSummary(ctx context.Context) (Summary, error) {
	queueInfo, err := c.inspector.GetQueueInfo(JudgeQueueName)
	if err != nil {
		if errors.Is(err, asynq.ErrQueueNotFound) {
			return Summary{
				Queue:       JudgeQueueName,
				RecentTasks: []TaskSummary{},
				UpdatedAt:   time.Now().UTC(),
			}, nil
		}
		return Summary{}, err
	}

	recentTasks := make([]TaskSummary, 0, 8)
	appendTasks := func(tasks []*asynq.TaskInfo, state string) {
		for _, task := range tasks {
			if len(recentTasks) >= 8 {
				return
			}
			recentTasks = append(recentTasks, taskSummaryFromInfo(task, state))
		}
	}

	completedTasks, err := c.inspector.ListCompletedTasks(JudgeQueueName, asynq.PageSize(4))
	if err == nil {
		appendTasks(completedTasks, "completed")
	}
	retryTasks, err := c.inspector.ListRetryTasks(JudgeQueueName, asynq.PageSize(4))
	if err == nil {
		appendTasks(retryTasks, "retry")
	}
	activeTasks, err := c.inspector.ListActiveTasks(JudgeQueueName, asynq.PageSize(2))
	if err == nil {
		appendTasks(activeTasks, "active")
	}
	pendingTasks, err := c.inspector.ListPendingTasks(JudgeQueueName, asynq.PageSize(2))
	if err == nil {
		appendTasks(pendingTasks, "pending")
	}

	return Summary{
		Queue:          queueInfo.Queue,
		Paused:         queueInfo.Paused,
		LatencySeconds: int(queueInfo.Latency.Round(time.Second).Seconds()),
		Pending:        queueInfo.Pending,
		Active:         queueInfo.Active,
		Scheduled:      queueInfo.Scheduled,
		Retry:          queueInfo.Retry,
		Archived:       queueInfo.Archived,
		Completed:      queueInfo.Completed,
		ProcessedToday: queueInfo.Processed,
		FailedToday:    queueInfo.Failed,
		RecentTasks:    recentTasks,
		UpdatedAt:      queueInfo.Timestamp.UTC(),
	}, nil
}

func taskSummaryFromInfo(task *asynq.TaskInfo, state string) TaskSummary {
	var payload JudgeSubmissionPayload
	_ = json.Unmarshal(task.Payload, &payload)
	return TaskSummary{
		ID:            task.ID,
		Type:          task.Type,
		SubmissionID:  submissionIDPtr(payload.SubmissionID),
		State:         state,
		CreatedAt:     nil,
		NextProcessAt: nullableTime(task.NextProcessAt),
		CompletedAt:   nullableTime(task.CompletedAt),
		LastErr:       task.LastErr,
	}
}

func submissionIDPtr(value uuid.UUID) *uuid.UUID {
	if value == uuid.Nil {
		return nil
	}
	submissionID := value
	return &submissionID
}

func nullableTime(value time.Time) *time.Time {
	if value.IsZero() {
		return nil
	}
	timestamp := value.UTC()
	return &timestamp
}
