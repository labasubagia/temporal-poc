package main

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"
	"path/filepath"

	"github.com/labasubagia/temporal-poc/internal"
	"go.temporal.io/api/enums/v1"
	"go.temporal.io/sdk/client"
)

type Server struct {
	temporal client.Client
}

func NewServer(temporal client.Client) *Server {
	return &Server{temporal: temporal}
}

func (s *Server) handleStartPayment(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	var req internal.PaymentRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	wo := client.StartWorkflowOptions{
		TaskQueue: "payment-worker",
		ID:        fmt.Sprintf("payment-%s-%d", req.OrderID, os.Getpid()),
	}

	run, err := s.temporal.ExecuteWorkflow(context.Background(), wo, "PaymentWorkflow", req)
	if err != nil {
		http.Error(w, fmt.Sprintf("failed to start workflow: %v", err), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]string{
		"workflow_id": run.GetID(),
		"run_id":      run.GetRunID(),
	})
}

func (s *Server) handleGetTimeline(w http.ResponseWriter, r *http.Request) {
	workflowID := r.URL.Query().Get("workflow_id")
	if workflowID == "" {
		http.Error(w, "workflow_id is required", http.StatusBadRequest)
		return
	}


	iter := s.temporal.GetWorkflowHistory(
		r.Context(), workflowID, "", false,
		enums.HISTORY_EVENT_FILTER_TYPE_ALL_EVENT,
	)

	result := internal.TimelineResponse{WorkflowID: workflowID}
	// scheduledTime keyed by scheduledEventID so we can correlate started/completed
	scheduledAt := map[int64]int64{}
	activityName := map[int64]string{}
	spans := map[int64]*internal.ActivitySpan{}
	totalActivities := 0
	completedActivities := 0

	for iter.HasNext() {
		event, err := iter.Next()
		if err != nil {
			http.Error(w, fmt.Sprintf("failed to read history: %v", err), http.StatusInternalServerError)
			return
		}

		ts := event.GetEventTime().AsTime().UnixMilli()
		switch event.GetEventType() {
		case enums.EVENT_TYPE_WORKFLOW_EXECUTION_STARTED:
			result.StartedAt = ts
		case enums.EVENT_TYPE_WORKFLOW_EXECUTION_COMPLETED,
			enums.EVENT_TYPE_WORKFLOW_EXECUTION_FAILED,
			enums.EVENT_TYPE_WORKFLOW_EXECUTION_TIMED_OUT:
			result.EndedAt = ts
		case enums.EVENT_TYPE_ACTIVITY_TASK_SCHEDULED:
			attrs := event.GetActivityTaskScheduledEventAttributes()
			scheduledAt[event.GetEventId()] = ts
			activityName[event.GetEventId()] = attrs.GetActivityType().GetName()
			totalActivities++
		case enums.EVENT_TYPE_ACTIVITY_TASK_STARTED:
			attrs := event.GetActivityTaskStartedEventAttributes()
			schedID := attrs.GetScheduledEventId()
			span := &internal.ActivitySpan{
				Name:      activityName[schedID],
				StartedAt: scheduledAt[schedID],
				Status:    "running",
			}
			spans[schedID] = span
			result.Activities = append(result.Activities, *span)
		case enums.EVENT_TYPE_ACTIVITY_TASK_COMPLETED:
			attrs := event.GetActivityTaskCompletedEventAttributes()
			schedID := attrs.GetScheduledEventId()
			if span, ok := spans[schedID]; ok {
				span.EndedAt = ts
				span.DurationMs = ts - span.StartedAt
				span.Status = "completed"
				completedActivities++
				// update the copy already in the slice
				for i := range result.Activities {
					if result.Activities[i].Name == span.Name && result.Activities[i].Status == "running" {
						result.Activities[i] = *span
						break
					}
				}
			}
		case enums.EVENT_TYPE_ACTIVITY_TASK_FAILED:
			attrs := event.GetActivityTaskFailedEventAttributes()
			schedID := attrs.GetScheduledEventId()
			if span, ok := spans[schedID]; ok {
				span.EndedAt = ts
				span.DurationMs = ts - span.StartedAt
				span.Status = "failed"
				completedActivities++
				for i := range result.Activities {
					if result.Activities[i].Name == span.Name && result.Activities[i].Status == "running" {
						result.Activities[i] = *span
						break
					}
				}
			}
		}
	}

	// Calculate progress as percentage of completed activities
	if totalActivities > 0 {
		result.Progress = (completedActivities * 100) / totalActivities
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(result)
}

func (s *Server) handleGetResult(w http.ResponseWriter, r *http.Request) {
	workflowID := r.URL.Query().Get("workflow_id")
	if workflowID == "" {
		http.Error(w, "workflow_id is required", http.StatusBadRequest)
		return
	}

	run := s.temporal.GetWorkflow(context.Background(), workflowID, "")

	var result internal.PaymentResult
	err := run.Get(context.Background(), &result)
	if err != nil {
		http.Error(w, fmt.Sprintf("workflow not completed: %v", err), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(result)
}

func main() {
	temporalAddr := os.Getenv("TEMPORAL_HOST_URL")
	if temporalAddr == "" {
		temporalAddr = "localhost:7233"
	}

	namespace := os.Getenv("TEMPORAL_NAMESPACE")
	if namespace == "" {
		namespace = "default"
	}

	c, err := client.NewLazyClient(client.Options{
		HostPort:  temporalAddr,
		Namespace: namespace,
	})
	if err != nil {
		log.Fatalf("unable to create Temporal client: %v", err)
	}
	defer c.Close()

	server := NewServer(c)

	exe, err := os.Executable()
	if err != nil {
		log.Fatal(err)
	}
	staticDir := filepath.Join(filepath.Dir(exe), "static")
	if _, err := os.Stat(staticDir); os.IsNotExist(err) {
		staticDir = "cmd/server/static"
	}

	mux := http.NewServeMux()
	mux.Handle("/static/", http.StripPrefix("/static/", http.FileServer(http.Dir(staticDir))))
	mux.HandleFunc("/api/payment/start", server.handleStartPayment)
	mux.HandleFunc("/api/payment/result", server.handleGetResult)
	mux.HandleFunc("/api/payment/timeline", server.handleGetTimeline)
	mux.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path == "/" {
			http.ServeFile(w, r, staticDir+"/index.html")
			return
		}
		http.NotFound(w, r)
	})

	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}

	log.Printf("Server starting on :%s", port)
	log.Fatal(http.ListenAndServe(":"+port, mux))
}