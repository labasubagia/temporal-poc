package workflow

import (
	"fmt"
	"time"

	"github.com/labasubagia/temporal-poc/activities"
	"github.com/labasubagia/temporal-poc/internal"
	"go.temporal.io/sdk/temporal"
	"go.temporal.io/sdk/workflow"
)

func PaymentWorkflow(ctx workflow.Context, req internal.PaymentRequest) (*internal.PaymentResult, error) {


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

	a := &activities.PaymentActivities{}

	if err := workflow.ExecuteActivity(ctx, a.ValidatePayment, req.OrderID, req.Amount).Get(ctx, nil); err != nil {
		return nil, fmt.Errorf("validate payment: %w", err)
	}

	transactionID := ""
	if err := workflow.ExecuteActivity(ctx, a.ProcessPayment, req.OrderID, req.Amount).Get(ctx, &transactionID); err != nil {
		return nil, fmt.Errorf("process payment: %w", err)
	}

	if err := workflow.ExecuteActivity(ctx, a.ConfirmPayment, transactionID).Get(ctx, nil); err != nil {
		return nil, fmt.Errorf("confirm payment: %w", err)
	}

	if err := workflow.ExecuteActivity(ctx, a.SendNotification, req.OrderID).Get(ctx, nil); err != nil {
		return nil, fmt.Errorf("send notification: %w", err)
	}


	return &internal.PaymentResult{
		TransactionID: transactionID,
		Status:        "completed",
		Message:       "Payment processed successfully",
	}, nil
}