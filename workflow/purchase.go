package workflow

import (
	"context"
	"time"

	"github.com/labasubagia/temporal-poc/activities"
	"github.com/labasubagia/temporal-poc/internal"
	"go.temporal.io/sdk/temporal"
	"go.temporal.io/sdk/workflow"
)

func PurchaseOrderWorkflow(ctx workflow.Context, req internal.PurchaseRequest) (*activities.PurchaseResult, error) {
	ao := workflow.ActivityOptions{
		StartToCloseTimeout: 2 * time.Minute,
		HeartbeatTimeout:   30 * time.Second,
		RetryPolicy: &temporal.RetryPolicy{
			InitialInterval:    time.Second,
			BackoffCoefficient: 2.0,
			MaximumInterval:    time.Minute,
			MaximumAttempts:    5,
		},
	}
	ctx = workflow.WithActivityOptions(ctx, ao)

	a := activities.NewPurchase()
	wfID := workflow.GetInfo(ctx).WorkflowExecution.ID

	notify := func(activity, status string) {
		_ = activities.NotifyProgress(context.Background(), wfID, activity, status)
	}

	notify("CreatePurchaseOrder", "started")
	if err := workflow.ExecuteActivity(ctx, a.CreatePurchaseOrder, req.OrderID, req.CustomerID, req.Items).Get(ctx, nil); err != nil {
		notify("CreatePurchaseOrder", "failed")
		return nil, err
	}
	notify("CreatePurchaseOrder", "completed")

	notify("ValidateStock", "started")
	if err := workflow.ExecuteActivity(ctx, a.ValidateStock, req.OrderID).Get(ctx, nil); err != nil {
		notify("ValidateStock", "failed")
		return nil, err
	}
	notify("ValidateStock", "completed")

	notify("AllocateItems", "started")
	if err := workflow.ExecuteActivity(ctx, a.AllocateItems, req.OrderID).Get(ctx, nil); err != nil {
		notify("AllocateItems", "failed")
		return nil, err
	}
	notify("AllocateItems", "completed")

	notify("CalculatePricing", "started")
	var amount float64
	if err := workflow.ExecuteActivity(ctx, a.CalculatePricing, req.OrderID).Get(ctx, &amount); err != nil {
		notify("CalculatePricing", "failed")
		return nil, err
	}
	notify("CalculatePricing", "completed")

	notify("ConfirmOrder", "started")
	if err := workflow.ExecuteActivity(ctx, a.ConfirmOrder, req.OrderID).Get(ctx, nil); err != nil {
		notify("ConfirmOrder", "failed")
		return nil, err
	}
	notify("ConfirmOrder", "completed")

	notify("NotifyCustomer", "started")
	if err := workflow.ExecuteActivity(ctx, a.NotifyCustomer, req.CustomerID).Get(ctx, nil); err != nil {
		notify("NotifyCustomer", "failed")
		return nil, err
	}
	notify("NotifyCustomer", "completed")

	notify("CompletePurchase", "started")
	var result *activities.PurchaseResult
	if err := workflow.ExecuteActivity(ctx, a.CompletePurchase, req.OrderID, amount).Get(ctx, &result); err != nil {
		notify("CompletePurchase", "failed")
		return nil, err
	}
	notify("CompletePurchase", "completed")

	return result, nil
}