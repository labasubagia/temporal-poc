package activities

import (
	"context"
	"errors"
	"os"
	"time"
)

type FailingActivities struct{}

func NewFailingActivities() *FailingActivities {
	return &FailingActivities{}
}

func getDelay() time.Duration {
	if v := os.Getenv("ACTIVITY_DELAY"); v != "" {
		if d, err := time.ParseDuration(v); err == nil {
			return d
		}
	}
	return 500 * time.Millisecond
}

func (a *FailingActivities) TaskOne(ctx context.Context, id string) error {
	time.Sleep(getDelay())
	return nil
}

func (a *FailingActivities) TaskTwo(ctx context.Context, id string) error {
	time.Sleep(getDelay())
	return errors.New("task two failed")
}

func (a *FailingActivities) TaskThree(ctx context.Context, id string) error {
	time.Sleep(getDelay())
	return nil
}

func (a *FailingActivities) TaskFour(ctx context.Context, id string) error {
	time.Sleep(getDelay())
	return nil
}