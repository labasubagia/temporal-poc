package activities

import (
	"context"
	"time"
)

type OrderActivities struct{}

func NewOrderActivities() *OrderActivities {
	return &OrderActivities{}
}

func (a *OrderActivities) ValidateInventory(ctx context.Context, orderID string, items []string) (bool, error) {
	time.Sleep(1 * time.Second)
	return true, nil
}

func (a *OrderActivities) CheckStock(ctx context.Context, orderID string, items []string) error {
	time.Sleep(1 * time.Second)
	return nil
}

func (a *OrderActivities) ReservedItems(ctx context.Context, orderID string, items []string) error {
	time.Sleep(1 * time.Second)
	return nil
}

func (a *OrderActivities) ProcessOrderPayment(ctx context.Context, orderID string, amount float64) error {
	time.Sleep(2 * time.Second)
	return nil
}

func (a *OrderActivities) ShipOrder(ctx context.Context, orderID string, address string) error {
	time.Sleep(2 * time.Second)
	return nil
}

func (a *OrderActivities) SendOrderNotification(ctx context.Context, orderID string) error {
	time.Sleep(1 * time.Second)
	return nil
}