# Design

## Overview

Payment and Order processing workflows using Temporal for reliable, observable, long-running operations.

## Architecture

```
cmd/server/       — HTTP server (API + static frontend)
cmd/worker/       — Temporal worker (registers workflows/activities)
activities/       — Activity implementations (non-deterministic)
workflow/         — Workflow definitions (deterministic)
internal/         — Shared types
```

## Frontend Routes

| Path | Description |
|------|-------------|
| `/` | Routing page - select Payment or Order workflow |
| `/payment` | Payment workflow UI |
| `/order` | Order workflow UI |

## Payment Workflow

### Flowchart

```mermaid
flowchart LR
    A[ValidatePayment<br/>2s] --> B[ProcessPayment<br/>3s]
    B --> C[ConfirmPayment<br/>2s]
    C --> D[SendNotification<br/>1s]
```

Total execution time: ~8 seconds.

### Activities

| Activity | Duration | Description |
|----------|----------|-------------|
| ValidatePayment | 2s | Validate order and payment details |
| ProcessPayment | 3s | Process payment with payment provider |
| ConfirmPayment | 2s | Confirm transaction |
| SendNotification | 1s | Send notification to customer |

### Activity Options

- StartToCloseTimeout: 2 minutes
- HeartbeatTimeout: 30 seconds
- Retry: Initial interval 1s, backoff 2x, max 5 attempts

## Order Fulfillment Workflow

### Flowchart

```mermaid
flowchart LR
    A[ValidateInventory<br/>1s] --> B[CheckStock<br/>1s]
    B --> C[ReservedItems<br/>1s]
    C --> D[ProcessOrderPayment<br/>2s]
    D --> E[ShipOrder<br/>2s]
    E --> F[SendOrderNotification<br/>1s]
```

Total execution time: ~8 seconds.

### Activities

| Activity | Duration | Description |
|----------|----------|-------------|
| ValidateInventory | 1s | Validate order and inventory |
| CheckStock | 1s | Check stock availability |
| ReservedItems | 1s | Reserve items in warehouse |
| ProcessOrderPayment | 2s | Process payment |
| ShipOrder | 2s | Ship order to customer |
| SendOrderNotification | 1s | Send notification to customer |

## API Endpoints

### POST /api/payment/start

Start a payment workflow.

Request:
```json
{
    "order_id": "ORD-123",
    "amount": 99.99,
    "customer_id": "CUST-456"
}
```

Response:
```json
{
    "workflow_id": "payment-ORD-123-xxx",
    "run_id": "xxx"
}
```

### POST /api/order/start

Start an order fulfillment workflow.

Request:
```json
{
    "order_id": "ORD-123",
    "customer_id": "CUST-456",
    "items": ["Item1", "Item2"]
}
```

Response:
```json
{
    "workflow_id": "order-ORD-123-xxx",
    "run_id": "xxx"
}
```

### GET /api/workflow/timeline?workflow_id=X

Get workflow timeline from history.

Response:
```json
{
    "workflow_id": "payment-ORD-123-xxx",
    "started_at_ms": 1778299327077,
    "ended_at_ms": 1778299335611,
    "progress": 100,
    "total_activities": 4,
    "activities": [...]
}
```

### GET /api/workflow/timeline-with-total-subprocess?workflow_id=X

Get workflow timeline with total activities from query handler.

Response:
```json
{
    "workflow_id": "order-ORD-123-xxx",
    "started_at_ms": 1778299301380,
    "ended_at_ms": 1778299317270,
    "progress": 100,
    "total_activities": 6,
    "activities": [...]
}
```

### GET /api/workflow/result?workflow_id=X

## Key Concepts

- **Workflow** — Deterministic execution, defines steps
- **Activity** — Non-deterministic operations (simulated with sleep)
- **Query handler** — Allows reading workflow state from outside without signals
- **Timeline API** — Reads workflow history to show activity progress

## Loading Progress

### Backend Implementation

The timeline API (`GET /api/workflow/timeline?workflow_id=X`) calculates progress based on completed activities.

```mermaid
flowchart TD
    A[Frontend Polls /timeline] --> B{expected_total?}
    B -->|true| C[Query workflow for total activities]
    B -->|false| D[Count scheduled from history]
    C --> E[progress = completed / total]
    D --> F[progress = completed / scheduled]
    E --> G[Return accurate progress]
    F --> H[Return estimate progress]
    G --> I[Frontend displays progress]
    H --> I
```

**With `expected_total=true`:**

The API queries the workflow via a query handler (`wf.QUERY_TOTAL_SUBPROCESS`) to get the total number of activities upfront. Progress is calculated as:

```
progress = (completed_activities * 100) / total_activities
```

This provides accurate progress from the start (e.g., 0%, 25%, 50%, 75%, 100%).

**Without `expected_total` (default):**

The API only knows activities that have been scheduled so far (from workflow history). Progress is calculated as:

```
progress = (completed_activities * 100) / scheduled_count
```

This is less accurate at the beginning because:
- First poll: 1 scheduled, 0 completed → 0%
- Second poll: 2 scheduled, 1 completed → 50%
- Third poll: 3 scheduled, 2 completed → 66%

So without `expected_total`, progress jumps to ~50% once the second activity starts.

See `cmd/server/main.go:206` (`handleGetWorkflowTimeline`) and `cmd/server/main.go:36` (`buildTimelineFromHistory`).

### Frontend Polling

```mermaid
sequenceDiagram
    participant U as User
    participant F as Frontend
    participant S as Server API
    participant T as Temporal

    U->>F: Click "Start Workflow"
    F->>S: POST /api/payment/start
    S->>T: Start workflow
    T-->>S: workflow_id
    S-->>F: {workflow_id: "..."}
    F->>S: GET /api/workflow/timeline?workflow_id=...&expected_total=true
    loop Every 1 second
        S->>T: Query workflow history
        T-->>S: Activity events
        S-->>F: {progress: X, activities: [...]}
        F->>F: Update UI
    end
    T-->>F: Workflow completes
    F->>F: Stop polling, show Complete
```

The frontend uses polling to fetch timeline/progress:

- `workflow-app.js` — Async/await polling loop with 1-second interval
- `runPollLoop()` fetches `/api/workflow/timeline` repeatedly until workflow completes or fails
- For workflows with known activity count (Payment, Order, Failing), `expected_total=true` is passed for accurate progress
- For dynamic/unknown workflows, default behavior is used

The polling approach ensures the UI stays updated without needing WebSockets or server push.

The frontend uses polling to fetch timeline/progress:

- `workflow-app.js` — Async/await polling loop with 1-second interval
- `runPollLoop()` fetches `/api/workflow/timeline` repeatedly until workflow completes or fails
- For workflows with known activity count (Payment, Order, Failing), `expected_total=true` is passed for accurate progress
- For dynamic/unknown workflows, default behavior is used

The polling approach ensures the UI stays updated without needing WebSockets or server push.
