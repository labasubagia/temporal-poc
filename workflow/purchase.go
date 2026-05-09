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

	// helper to execute an activity and send start/failed/completed notifications
	withNotify := func(fn any, out any, args ...any) error {
		_ = activities.NotifyProgress(context.Background(), wfID, fn, activities.STATUS_STARTED, totalActivities)
		if err := workflow.ExecuteActivity(ctx, fn, args...).Get(ctx, out); err != nil {
			_ = activities.NotifyProgress(context.Background(), wfID, fn, activities.STATUS_FAILED, totalActivities)
			return err
		}
		_ = activities.NotifyProgress(context.Background(), wfID, fn, activities.STATUS_COMPLETED, totalActivities)
		return nil
	}

	if err := withNotify(a.CreatePurchaseOrder, nil, req.OrderID, req.CustomerID, req.Items); err != nil {
		return nil, err
	}

	if err := withNotify(a.ValidateStock, nil, req.OrderID); err != nil {
		return nil, err
	}

	if err := withNotify(a.AllocateItems, nil, req.OrderID); err != nil {
		return nil, err
	}

	var amount float64
	if err := withNotify(a.CalculatePricing, &amount, req.OrderID); err != nil {
		return nil, err
	}

	if err := withNotify(a.ConfirmOrder, nil, req.OrderID); err != nil {
		return nil, err
	}

	if err := withNotify(a.NotifyCustomer, nil, req.CustomerID); err != nil {
		return nil, err
	}

	var result *activities.PurchaseResult
	if err := withNotify(a.CompletePurchase, &result, req.OrderID, amount); err != nil {
		return nil, err
	}

	return result, nil
}