# Temporal PoC

A proof-of-concept demonstrating long-running payment processing with Temporal, featuring a frontend that displays workflow progress in real-time.

## Architecture

```
cmd/server/       — HTTP server (API + static frontend)
cmd/worker/       — Temporal worker (registers workflows/activities)
cmd/ws/           — WebSocket server (real-time updates)
activities/       — Activity implementations (non-deterministic)
workflow/         — Workflow definitions (deterministic)
internal/         — Shared types
```

## Frontend Routes

| Path | Description |
|------|-------------|
| `/` | Routing page - select Payment or Order workflow |
| `/payment` | Payment workflow UI |
| `/order` | Order fulfillment workflow UI |
| `/ws-purchase` | Purchase Order UI (WebSocket real-time) |
| `/failing` | Failing workflow UI |

## Workflows

### Payment Workflow (~8s)
`ValidatePayment (2s) → ProcessPayment (3s) → ConfirmPayment (2s) → SendNotification (1s)`

### Order Fulfillment Workflow (~8s)
`ValidateInventory → CheckStock → ReservedItems → ProcessOrderPayment → ShipOrder → SendOrderNotification`

### Purchase Order Workflow (~10s, WebSocket real-time)
`CreatePurchaseOrder → ValidateStock → AllocateItems → CalculatePricing → ConfirmOrder → NotifyCustomer → CompletePurchase`

## Progress Updates

| Approach | Workflows | Mechanism |
|----------|-----------|-----------|
| **Polling** | Payment, Order, Failing | Frontend polls `/api/workflow/timeline` every 1s |
| **WebSocket** | Purchase Order | Real-time updates via ws-server |

See [docs/design.md](docs/design.md) for detailed documentation.

## Demo

<video src="https://github.com/user-attachments/assets/abda8ed6-10b1-4ea6-8ffa-d6f44b44a371"></video>

## How it works

**Payment, Order, Failing workflows (Polling):**
1. Frontend calls `POST /api/payment/start` with workflow parameters
2. Server starts a Temporal workflow and returns `workflow_id`
3. Frontend polls `GET /api/workflow/timeline?workflow_id=X&expected_total=true` every 1s
4. Progress calculated as `(completed_activities * 100) / total_activities`

**Purchase Order workflow (WebSocket real-time):**
1. Frontend calls `POST /api/purchase/start`
2. Server starts workflow and returns `workflow_id`
3. Frontend connects to `ws://localhost:8081/ws?workflow_id=X`
4. Worker calls `NotifyProgress` activity after each step
5. ws-server broadcasts updates to connected clients in real-time

### Key concepts

- **Workflow** — Deterministic execution, defines steps as code
- **Activity** — Non-deterministic operations (DB, HTTP, file I/O)
- **Query handler** — Allows reading workflow state from outside

## Setup

```sh
cp .env.example .env
```

## Running

### Option 1: Docker Compose (Recommended for development)

```sh
podman compose -f compose-dev.yml up -d
```

This starts all services with live-reload:
- **temporal**: Temporal server on port 7233
- **server**: HTTP server on port 8080
- **worker**: Temporal worker
- **ws-server**: WebSocket server on port 8081

### Option 2: Manual

### 1. Start Temporal

```sh
podman compose up -d
```

### 2. Start worker

```sh
air -c air-worker.toml
```

### 3. Start server

```sh
air -c air-server.toml
```

### 4. Start WebSocket server

```sh
air -c air-ws.toml
```

### 5. Open browser

```
http://localhost:8080
```

## Commands

```sh
# Run from project root directory
air -c air-worker.toml            # Start Temporal worker (dev)
air -c air-server.toml           # Start HTTP server (dev)
air -c air-ws.toml               # Start WebSocket server (dev)

go run ./cmd/worker              # Start Temporal worker (prod)
go run ./cmd/server             # Start HTTP server (prod)
go run ./cmd/ws                  # Start WebSocket server (prod)

go test ./...                    # Run tests
golangci-lint run                # Lint
go build ./...                   # Build
```