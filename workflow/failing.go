package workflow

import (
	"fmt"
	"time"

	"github.com/labasubagia/temporal-poc/activities"
	"github.com/labasubagia/temporal-poc/internal"
	"go.temporal.io/sdk/temporal"
	"go.temporal.io/sdk/workflow"
)

func FailingWorkflow(ctx workflow.Context, req internal.FailingRequest) (string, error) {

	totalActivities := 4

	ao := workflow.ActivityOptions{
		StartToCloseTimeout: 2 * time.Minute,
		HeartbeatTimeout:    30 * time.Second,
		RetryPolicy: &temporal.RetryPolicy{
			InitialInterval:    time.Second,
			BackoffCoefficient: 2.0,
			MaximumInterval:    time.Minute,
			MaximumAttempts:    1,
		},
	}
	ctx = workflow.WithActivityOptions(ctx, ao)

	err := workflow.SetQueryHandler(ctx, QUERY_TOTAL_SUBPROCESS, func() (int, error) {
		return totalActivities, nil
	})
	if err != nil {
		return "", fmt.Errorf("set query handler: %w", err)
	}

	a := &activities.FailingActivities{}

	err = workflow.ExecuteActivity(ctx, a.TaskOne, req.ID).Get(ctx, nil)
	if err != nil {
		return "", fmt.Errorf("task one: %w", err)
	}

	err = workflow.ExecuteActivity(ctx, a.TaskTwo, req.ID).Get(ctx, nil)
	if err != nil {
		return "", fmt.Errorf("task two: %w", err)
	}

	err = workflow.ExecuteActivity(ctx, a.TaskThree, req.ID).Get(ctx, nil)
	if err != nil {
		return "", fmt.Errorf("task three: %w", err)
	}

	err = workflow.ExecuteActivity(ctx, a.TaskFour, req.ID).Get(ctx, nil)
	if err != nil {
		return "", fmt.Errorf("task four: %w", err)
	}

	return "completed", nil
}