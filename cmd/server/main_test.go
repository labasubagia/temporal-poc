package main

import (
	"testing"
	"time"

	"github.com/labasubagia/temporal-poc/internal"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	commonpb "go.temporal.io/api/common/v1"
	enums "go.temporal.io/api/enums/v1"
	historypb "go.temporal.io/api/history/v1"
	"google.golang.org/protobuf/types/known/timestamppb"
)

type stubHistoryEventIterator struct {
	events []*historypb.HistoryEvent
	index  int
}

func (s *stubHistoryEventIterator) HasNext() bool {
	return s.index < len(s.events)
}

func (s *stubHistoryEventIterator) Next() (*historypb.HistoryEvent, error) {
	event := s.events[s.index]
	s.index++
	return event, nil
}

func testHistoryEvent(eventID int64, eventType enums.EventType, ts time.Time, attrs any) *historypb.HistoryEvent {
	event := &historypb.HistoryEvent{
		EventId:   eventID,
		EventTime: timestamppb.New(ts),
		EventType: eventType,
	}

	switch attr := attrs.(type) {
	case *historypb.WorkflowExecutionStartedEventAttributes:
		event.Attributes = &historypb.HistoryEvent_WorkflowExecutionStartedEventAttributes{WorkflowExecutionStartedEventAttributes: attr}
	case *historypb.WorkflowExecutionCompletedEventAttributes:
		event.Attributes = &historypb.HistoryEvent_WorkflowExecutionCompletedEventAttributes{WorkflowExecutionCompletedEventAttributes: attr}
	case *historypb.ActivityTaskScheduledEventAttributes:
		event.Attributes = &historypb.HistoryEvent_ActivityTaskScheduledEventAttributes{ActivityTaskScheduledEventAttributes: attr}
	case *historypb.ActivityTaskStartedEventAttributes:
		event.Attributes = &historypb.HistoryEvent_ActivityTaskStartedEventAttributes{ActivityTaskStartedEventAttributes: attr}
	case *historypb.ActivityTaskCompletedEventAttributes:
		event.Attributes = &historypb.HistoryEvent_ActivityTaskCompletedEventAttributes{ActivityTaskCompletedEventAttributes: attr}
	case *historypb.ActivityTaskFailedEventAttributes:
		event.Attributes = &historypb.HistoryEvent_ActivityTaskFailedEventAttributes{ActivityTaskFailedEventAttributes: attr}
	case nil:
		return event
	default:
		panic("unsupported test history attributes")
	}

	return event
}

func TestCalculateProgress(t *testing.T) {
	tests := []struct {
		name      string
		completed int
		scheduled int
		query    int
		expected int
	}{
		{"zero both", 0, 0, 0, 0},
		{"zero scheduled", 0, 0, 4, 0},
		{"all completed", 4, 4, 0, 100},
		{"half completed", 2, 4, 0, 50},
		{"one quarter", 1, 4, 0, 25},
		{"three quarters", 3, 4, 0, 75},
		{"six of six", 6, 6, 0, 100},
		{"one of six", 1, 6, 0, 16},
		{"query total", 1, 0, 4, 25},
		{"query overrides", 1, 4, 4, 25},
		{"query takes precedence", 2, 3, 4, 50},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := calculateProgress(tt.completed, tt.scheduled, tt.query)
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestBuildTimelineFromHistory_UsesQueryTotalForProgress(t *testing.T) {
	base := time.UnixMilli(1_000)
	iter := &stubHistoryEventIterator{events: []*historypb.HistoryEvent{
		testHistoryEvent(1, enums.EVENT_TYPE_WORKFLOW_EXECUTION_STARTED, base, &historypb.WorkflowExecutionStartedEventAttributes{}),
		testHistoryEvent(2, enums.EVENT_TYPE_ACTIVITY_TASK_SCHEDULED, base.Add(100*time.Millisecond), &historypb.ActivityTaskScheduledEventAttributes{
			ActivityType: &commonpb.ActivityType{Name: "ChargeCard"},
		}),
		testHistoryEvent(3, enums.EVENT_TYPE_ACTIVITY_TASK_STARTED, base.Add(200*time.Millisecond), &historypb.ActivityTaskStartedEventAttributes{
			ScheduledEventId: 2,
		}),
		testHistoryEvent(4, enums.EVENT_TYPE_ACTIVITY_TASK_COMPLETED, base.Add(800*time.Millisecond), &historypb.ActivityTaskCompletedEventAttributes{
			ScheduledEventId: 2,
		}),
		testHistoryEvent(5, enums.EVENT_TYPE_WORKFLOW_EXECUTION_COMPLETED, base.Add(900*time.Millisecond), &historypb.WorkflowExecutionCompletedEventAttributes{}),
	}}

	result := buildTimelineFromHistory(iter, 4, true)

	assert.Equal(t, base.UnixMilli(), result.StartedAt)
	assert.Equal(t, base.Add(900*time.Millisecond).UnixMilli(), result.EndedAt)
	assert.Equal(t, 4, result.TotalActivities)
	assert.Equal(t, 25, result.Progress)
	require.Len(t, result.Activities, 1)
	assert.Equal(t, internal.ActivitySpan{
		Name:       "ChargeCard",
		StartedAt:  base.Add(100 * time.Millisecond).UnixMilli(),
		EndedAt:    base.Add(800 * time.Millisecond).UnixMilli(),
		DurationMs: 700,
		Status:     "completed",
	}, result.Activities[0])
}

func TestBuildTimelineFromHistory_MarksFailedActivities(t *testing.T) {
	base := time.UnixMilli(10_000)
	iter := &stubHistoryEventIterator{events: []*historypb.HistoryEvent{
		testHistoryEvent(1, enums.EVENT_TYPE_WORKFLOW_EXECUTION_STARTED, base, &historypb.WorkflowExecutionStartedEventAttributes{}),
		testHistoryEvent(2, enums.EVENT_TYPE_ACTIVITY_TASK_SCHEDULED, base.Add(50*time.Millisecond), &historypb.ActivityTaskScheduledEventAttributes{
			ActivityType: &commonpb.ActivityType{Name: "ReserveInventory"},
		}),
		testHistoryEvent(3, enums.EVENT_TYPE_ACTIVITY_TASK_STARTED, base.Add(100*time.Millisecond), &historypb.ActivityTaskStartedEventAttributes{
			ScheduledEventId: 2,
		}),
		testHistoryEvent(4, enums.EVENT_TYPE_ACTIVITY_TASK_FAILED, base.Add(250*time.Millisecond), &historypb.ActivityTaskFailedEventAttributes{
			ScheduledEventId: 2,
		}),
	}}

	result := buildTimelineFromHistory(iter, 0, false)

	assert.Equal(t, 1, result.TotalActivities)
	assert.Equal(t, 0, result.Progress)
	require.Len(t, result.Activities, 1)
	assert.Equal(t, internal.ActivitySpan{
		Name:       "ReserveInventory",
		StartedAt:  base.Add(50 * time.Millisecond).UnixMilli(),
		EndedAt:    base.Add(250 * time.Millisecond).UnixMilli(),
		DurationMs: 200,
		Status:     "failed",
	}, result.Activities[0])
}

func TestBuildTimelineFromHistory_UsesScheduledCountWhenExpectedTotalDisabled(t *testing.T) {
	base := time.UnixMilli(20_000)
	iter := &stubHistoryEventIterator{events: []*historypb.HistoryEvent{
		testHistoryEvent(1, enums.EVENT_TYPE_WORKFLOW_EXECUTION_STARTED, base, &historypb.WorkflowExecutionStartedEventAttributes{}),
		testHistoryEvent(2, enums.EVENT_TYPE_ACTIVITY_TASK_SCHEDULED, base.Add(50*time.Millisecond), &historypb.ActivityTaskScheduledEventAttributes{
			ActivityType: &commonpb.ActivityType{Name: "ValidateOrder"},
		}),
		testHistoryEvent(3, enums.EVENT_TYPE_ACTIVITY_TASK_STARTED, base.Add(100*time.Millisecond), &historypb.ActivityTaskStartedEventAttributes{
			ScheduledEventId: 2,
		}),
		testHistoryEvent(4, enums.EVENT_TYPE_ACTIVITY_TASK_COMPLETED, base.Add(200*time.Millisecond), &historypb.ActivityTaskCompletedEventAttributes{
			ScheduledEventId: 2,
		}),
		testHistoryEvent(5, enums.EVENT_TYPE_ACTIVITY_TASK_SCHEDULED, base.Add(250*time.Millisecond), &historypb.ActivityTaskScheduledEventAttributes{
			ActivityType: &commonpb.ActivityType{Name: "NotifyWarehouse"},
		}),
		testHistoryEvent(6, enums.EVENT_TYPE_ACTIVITY_TASK_STARTED, base.Add(300*time.Millisecond), &historypb.ActivityTaskStartedEventAttributes{
			ScheduledEventId: 5,
		}),
		testHistoryEvent(7, enums.EVENT_TYPE_WORKFLOW_EXECUTION_COMPLETED, base.Add(400*time.Millisecond), &historypb.WorkflowExecutionCompletedEventAttributes{}),
	}}

	result := buildTimelineFromHistory(iter, 99, false)

	assert.Equal(t, 2, result.TotalActivities)
	assert.Equal(t, 50, result.Progress)
	require.Len(t, result.Activities, 2)
	assert.Equal(t, internal.ActivitySpan{
		Name:       "ValidateOrder",
		StartedAt:  base.Add(50 * time.Millisecond).UnixMilli(),
		EndedAt:    base.Add(200 * time.Millisecond).UnixMilli(),
		DurationMs: 150,
		Status:     "completed",
	}, result.Activities[0])
	assert.Equal(t, internal.ActivitySpan{
		Name:      "NotifyWarehouse",
		StartedAt: base.Add(250 * time.Millisecond).UnixMilli(),
		Status:    "running",
	}, result.Activities[1])
}

func TestBuildTimelineFromHistory_SetsEndTimeOnWorkflowFailure(t *testing.T) {
	base := time.UnixMilli(30_000)
	iter := &stubHistoryEventIterator{events: []*historypb.HistoryEvent{
		testHistoryEvent(1, enums.EVENT_TYPE_WORKFLOW_EXECUTION_STARTED, base, &historypb.WorkflowExecutionStartedEventAttributes{}),
		testHistoryEvent(2, enums.EVENT_TYPE_ACTIVITY_TASK_SCHEDULED, base.Add(25*time.Millisecond), &historypb.ActivityTaskScheduledEventAttributes{
			ActivityType: &commonpb.ActivityType{Name: "ChargeCard"},
		}),
		testHistoryEvent(3, enums.EVENT_TYPE_ACTIVITY_TASK_STARTED, base.Add(50*time.Millisecond), &historypb.ActivityTaskStartedEventAttributes{
			ScheduledEventId: 2,
		}),
		testHistoryEvent(4, enums.EVENT_TYPE_ACTIVITY_TASK_FAILED, base.Add(175*time.Millisecond), &historypb.ActivityTaskFailedEventAttributes{
			ScheduledEventId: 2,
		}),
		testHistoryEvent(5, enums.EVENT_TYPE_WORKFLOW_EXECUTION_FAILED, base.Add(300*time.Millisecond), nil),
	}}

	result := buildTimelineFromHistory(iter, 0, false)

	assert.Equal(t, base.UnixMilli(), result.StartedAt)
	assert.Equal(t, base.Add(300*time.Millisecond).UnixMilli(), result.EndedAt)
	assert.Equal(t, 1, result.TotalActivities)
	assert.Equal(t, 0, result.Progress)
	require.Len(t, result.Activities, 1)
	assert.Equal(t, "failed", result.Activities[0].Status)
}

func TestBuildTimelineFromHistory_SetsEndTimeOnWorkflowTimeout(t *testing.T) {
	base := time.UnixMilli(40_000)
	iter := &stubHistoryEventIterator{events: []*historypb.HistoryEvent{
		testHistoryEvent(1, enums.EVENT_TYPE_WORKFLOW_EXECUTION_STARTED, base, &historypb.WorkflowExecutionStartedEventAttributes{}),
		testHistoryEvent(2, enums.EVENT_TYPE_ACTIVITY_TASK_SCHEDULED, base.Add(10*time.Millisecond), &historypb.ActivityTaskScheduledEventAttributes{
			ActivityType: &commonpb.ActivityType{Name: "SendReceipt"},
		}),
		testHistoryEvent(3, enums.EVENT_TYPE_ACTIVITY_TASK_STARTED, base.Add(20*time.Millisecond), &historypb.ActivityTaskStartedEventAttributes{
			ScheduledEventId: 2,
		}),
		testHistoryEvent(4, enums.EVENT_TYPE_WORKFLOW_EXECUTION_TIMED_OUT, base.Add(500*time.Millisecond), nil),
	}}

	result := buildTimelineFromHistory(iter, 0, false)

	assert.Equal(t, base.UnixMilli(), result.StartedAt)
	assert.Equal(t, base.Add(500*time.Millisecond).UnixMilli(), result.EndedAt)
	assert.Equal(t, 1, result.TotalActivities)
	assert.Equal(t, 0, result.Progress)
	require.Len(t, result.Activities, 1)
	assert.Equal(t, internal.ActivitySpan{
		Name:      "SendReceipt",
		StartedAt: base.Add(10 * time.Millisecond).UnixMilli(),
		Status:    "running",
	}, result.Activities[0])
}