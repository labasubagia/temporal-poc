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
	Name        string  `json:"name"`
	StartedAt  int64   `json:"started_at_ms"`
	EndedAt    int64   `json:"ended_at_ms"`
	DurationMs int64  `json:"duration_ms"`
	Status    string  `json:"status"`
}

type TimelineResponse struct {
	WorkflowID      string         `json:"workflow_id"`
	StartedAt      int64          `json:"started_at_ms"`
	EndedAt        int64          `json:"ended_at_ms"`
	Progress       int            `json:"progress"`
	TotalActivities int            `json:"total_activities"`
	Activities     []ActivitySpan `json:"activities"`
}

type OrderRequest struct {
	OrderID     string `json:"order_id"`
	CustomerID  string `json:"customer_id"`
	Items       []string `json:"items"`
}

type OrderProgress struct {
	Progress int    `json:"progress"`
	Step     string `json:"step"`
	Complete bool   `json:"complete"`
}

type FailingRequest struct {
	ID string `json:"id"`
}