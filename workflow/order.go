package workflow

import (
	"fmt"
	"time"

	"github.com/labasubagia/temporal-poc/activities"
	"github.com/labasubagia/temporal-poc/internal"
	"go.temporal.io/sdk/temporal"
	"go.temporal.io/sdk/workflow"
)

func OrderFulfillmentWorkflow(ctx workflow.Context, req internal.OrderRequest) (string, error) {


	ao := workflow.ActivityOptions{
		StartToCloseTimeout: 2 * time.Minute,
		HeartbeatTimeout:    30 * time.Second,
		RetryPolicy: &temporal.RetryPolicy{
			InitialInterval:    time.Second,
			BackoffCoefficient: 2.0,
			MaximumInterval:    time.Minute,
			MaximumAttempts:    5,
		},
	}
	ctx = workflow.WithActivityOptions(ctx, ao)

	
	// total activities
	totalSubprocess := 6

	err := workflow.SetQueryHandler(ctx, QUERY_TOTAL_SUBPROCESS, func() (int, error) {
		return totalSubprocess, nil
	})
	if err != nil {
		return "", fmt.Errorf("set query handler: %w", err)
	}

	a := &activities.OrderActivities{}

	err = workflow.ExecuteActivity(ctx, a.ValidateInventory, req.OrderID, req.Items).Get(ctx, nil)
	if err != nil {
		return "", fmt.Errorf("validate inventory: %w", err)
	}

	err = workflow.ExecuteActivity(ctx, a.CheckStock, req.OrderID, req.Items).Get(ctx, nil)
	if err != nil {
		return "", fmt.Errorf("check stock: %w", err)
	}

	err = workflow.ExecuteActivity(ctx, a.ReservedItems, req.OrderID, req.Items).Get(ctx, nil)
	if err != nil {
		return "", fmt.Errorf("reserve items: %w", err)
	}

	err = workflow.ExecuteActivity(ctx, a.ProcessOrderPayment, req.OrderID, 100.0).Get(ctx, nil)
	if err != nil {
		return "", fmt.Errorf("process payment: %w", err)
	}

	err = workflow.ExecuteActivity(ctx, a.ShipOrder, req.OrderID, "123 Main St").Get(ctx, nil)
	if err != nil {
		return "", fmt.Errorf("ship order: %w", err)
	}

	err = workflow.ExecuteActivity(ctx, a.SendOrderNotification, req.OrderID).Get(ctx, nil)
	if err != nil {
		return "", fmt.Errorf("send notification: %w", err)
	}

	return fmt.Sprintf("order-%s-completed", req.OrderID), nil
}