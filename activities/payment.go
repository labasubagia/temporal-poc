package activities

import (
	"context"
	"fmt"
	"time"
)

type PaymentActivities struct{}

func New() *PaymentActivities {
	return &PaymentActivities{}
}

func (a *PaymentActivities) ValidatePayment(ctx context.Context, orderID string, amount float64) error {
	time.Sleep(2 * time.Second)
	return nil
}

func (a *PaymentActivities) ProcessPayment(ctx context.Context, orderID string, amount float64) (string, error) {
	time.Sleep(3 * time.Second)
	return fmt.Sprintf("TXN-%d", time.Now().UnixNano()), nil
}

func (a *PaymentActivities) ConfirmPayment(ctx context.Context, transactionID string) error {
	time.Sleep(2 * time.Second)
	return nil
}

func (a *PaymentActivities) SendNotification(ctx context.Context, orderID string) error {
	time.Sleep(1 * time.Second)
	return nil
}