package activities

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"time"
)

type PurchaseActivities struct{}

func NewPurchase() *PurchaseActivities {
	return &PurchaseActivities{}
}

func (a *PurchaseActivities) CreatePurchaseOrder(ctx context.Context, orderID, customerID string, items []string) error {
	time.Sleep(1 * time.Second)
	return nil
}

func (a *PurchaseActivities) ValidateStock(ctx context.Context, orderID string) error {
	time.Sleep(1 * time.Second)
	return nil
}

func (a *PurchaseActivities) AllocateItems(ctx context.Context, orderID string) error {
	time.Sleep(2 * time.Second)
	return nil
}

func (a *PurchaseActivities) CalculatePricing(ctx context.Context, orderID string) (float64, error) {
	time.Sleep(1 * time.Second)
	return 99.99, nil
}

func (a *PurchaseActivities) ConfirmOrder(ctx context.Context, orderID string) error {
	time.Sleep(1 * time.Second)
	return nil
}

func (a *PurchaseActivities) NotifyCustomer(ctx context.Context, customerID string) error {
	time.Sleep(1 * time.Second)
	return nil
}

type PurchaseResult struct {
	OrderID     string  `json:"order_id"`
	TotalAmount float64 `json:"total_amount"`
	Status      string  `json:"status"`
	Message     string  `json:"message"`
}

func (a *PurchaseActivities) CompletePurchase(ctx context.Context, orderID string, amount float64) (*PurchaseResult, error) {
	time.Sleep(1 * time.Second)
	return &PurchaseResult{
		OrderID:     orderID,
		TotalAmount: amount,
		Status:      "completed",
		Message:     fmt.Sprintf("Order %s completed successfully", orderID),
	}, nil
}

func NotifyProgress(ctx context.Context, workflowID, activityName, status string) error {
	payload := map[string]string{
		"workflow_id": workflowID,
		"activity":   activityName,
		"status":     status,
	}
	body, _ := json.Marshal(payload)
	_, _ = http.Post("http://localhost:8081/ws/notify", "application/json", bytes.NewReader(body))
	return nil
}