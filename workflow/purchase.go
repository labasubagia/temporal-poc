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

	totalActivities := 7

	notify := func(fn any, status string) {
		_ = activities.NotifyProgress(context.Background(), wfID, fn, status, totalActivities)
	}

	notify(a.CreatePurchaseOrder, activities.STATUS_STARTED)
	if err := workflow.ExecuteActivity(ctx, a.CreatePurchaseOrder, req.OrderID, req.CustomerID, req.Items).Get(ctx, nil); err != nil {
		notify(a.CreatePurchaseOrder, activities.STATUS_FAILED)
		return nil, err
	}
	notify(a.CreatePurchaseOrder, activities.STATUS_COMPLETED)

	notify(a.ValidateStock, activities.STATUS_STARTED)
	if err := workflow.ExecuteActivity(ctx, a.ValidateStock, req.OrderID).Get(ctx, nil); err != nil {
		notify(a.ValidateStock, activities.STATUS_FAILED)
		return nil, err
	}
	notify(a.ValidateStock, activities.STATUS_COMPLETED)

	notify(a.AllocateItems, activities.STATUS_STARTED)
	if err := workflow.ExecuteActivity(ctx, a.AllocateItems, req.OrderID).Get(ctx, nil); err != nil {
		notify(a.AllocateItems, activities.STATUS_FAILED)
		return nil, err
	}
	notify(a.AllocateItems, activities.STATUS_COMPLETED)

	notify(a.CalculatePricing, activities.STATUS_STARTED)
	var amount float64
	if err := workflow.ExecuteActivity(ctx, a.CalculatePricing, req.OrderID).Get(ctx, &amount); err != nil {
		notify(a.CalculatePricing, activities.STATUS_FAILED)
		return nil, err
	}
	notify(a.CalculatePricing, activities.STATUS_COMPLETED)

	notify(a.ConfirmOrder, activities.STATUS_STARTED)
	if err := workflow.ExecuteActivity(ctx, a.ConfirmOrder, req.OrderID).Get(ctx, nil); err != nil {
		notify(a.ConfirmOrder, activities.STATUS_FAILED)
		return nil, err
	}
	notify(a.ConfirmOrder, activities.STATUS_COMPLETED)

	notify(a.NotifyCustomer, activities.STATUS_STARTED)
	if err := workflow.ExecuteActivity(ctx, a.NotifyCustomer, req.CustomerID).Get(ctx, nil); err != nil {
		notify(a.NotifyCustomer, activities.STATUS_FAILED)
		return nil, err
	}
	notify(a.NotifyCustomer, activities.STATUS_COMPLETED)

	notify(a.CompletePurchase, activities.STATUS_STARTED)
	var result *activities.PurchaseResult
	if err := workflow.ExecuteActivity(ctx, a.CompletePurchase, req.OrderID, amount).Get(ctx, &result); err != nil {
		notify(a.CompletePurchase, activities.STATUS_FAILED)
		return nil, err
	}
	notify(a.CompletePurchase, activities.STATUS_COMPLETED)

	return result, nil
}