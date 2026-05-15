package judge

import (
	"context"
	"fmt"
)

type NotImplementedExecutor struct{}

func (NotImplementedExecutor) Run(ctx context.Context, submission Submission, testCase TestCase) (RunResult, error) {
	return RunResult{}, fmt.Errorf("judge execution is not implemented yet")
}
