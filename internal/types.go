package internal

type PaymentRequest struct {
	OrderID    string  `json:"order_id"`
	Amount     float64 `json:"amount"`
	CustomerID string  `json:"customer_id"`
}

type PaymentResult struct {
	TransactionID string `json:"transaction_id"`
	Status        string `json:"status"`
	Message       string `json:"message"`
}

type WorkflowStatus struct {
	WorkflowID string `json:"workflow_id"`
	Progress   int    `json:"progress"`
	Step       string `json:"step"`
	Activity   string `json:"activity"`
	Complete   bool   `json:"complete"`
}

type ProgressQuery struct{}

type ActivitySpan struct {
	Name      string  `json:"name"`
	StartedAt int64   `json:"started_at_ms"`  // unix ms
	EndedAt   int64   `json:"ended_at_ms"`    // unix ms, 0 if still running
	DurationMs int64  `json:"duration_ms"`    // 0 if still running
	Status    string  `json:"status"`         // "running", "completed", "failed"
}

type TimelineResponse struct {
	WorkflowID  string         `json:"workflow_id"`
	StartedAt   int64          `json:"started_at_ms"`
	EndedAt     int64          `json:"ended_at_ms"`
	Progress    int            `json:"progress"`
	Activities  []ActivitySpan `json:"activities"`
}