package activities

import (
	"context"
	"errors"
	"time"
)

type FailingActivities struct{}

func NewFailingActivities() *FailingActivities {
	return &FailingActivities{}
}


func (a *FailingActivities) TaskOne(ctx context.Context, id string) error {
	time.Sleep(1 * time.Second)
	return nil
}

func (a *FailingActivities) TaskTwo(ctx context.Context, id string) error {
	time.Sleep(1 * time.Second)
	return errors.New("task two failed")
}

func (a *FailingActivities) TaskThree(ctx context.Context, id string) error {
	time.Sleep(1 * time.Second)
	return nil
}

func (a *FailingActivities) TaskFour(ctx context.Context, id string) error {
	time.Sleep(1 * time.Second)
	return nil
}