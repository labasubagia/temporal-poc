# Temporal PoC

A proof-of-concept demonstrating long-running payment processing with Temporal, featuring a frontend that displays workflow progress in real-time.

## Architecture

```
cmd/server/       — HTTP server (API + static frontend)
cmd/worker/       — Temporal worker (registers workflows/activities)
activities/       — Activity implementations (non-deterministic)
workflow/         — Workflow definitions (deterministic)
internal/         — Shared types
```

## How it works

1. Frontend calls `POST /api/payment/start` with payment details
2. Server starts a Temporal workflow and returns `workflow_id`
3. Frontend polls `GET /api/payment/progress?workflow_id=X` every 500ms
4. Progress updates display in the UI with percentage

### Key concepts

- **Workflow** — Defines steps as deterministic execution
- **Activity** — Non-deterministic operations (DB, HTTP, etc.)
- **Query handler** — Allows reading workflow state from outside

## Setup

```sh
cp .env.example .env
```

## Running

### 1. Start Temporal

```sh
podman compose up -d
```

### 2. Start worker (development)

```sh
air -c air-worker.toml
```

### 3. Start server (development)

```sh
air -c air-server.toml
```

### 4. Open browser

```
http://localhost:8080
```

## Commands

```sh
# Run from project root directory
air -c air-worker.toml            # Start Temporal worker (dev)
air -c air-server.toml           # Start HTTP server (dev)

go run ./cmd/worker              # Start Temporal worker (prod)
go run ./cmd/server             # Start HTTP server (prod)

go test ./...                    # Run tests
golangci-lint run                # Lint
go build ./...                   # Build
```