# Agent Instructions

See [README.md](README.md) for project documentation and setup.
See [docs/](docs/) for design documentation.

## Before commit

Run lint before committing:

```sh
golangci-lint run
```

## Quick commands

```sh
air -c air-worker.toml          # Start Temporal worker (dev)
air -c air-server.toml          # Start frontend/server (dev)
air -c air-ws.toml              # Start WebSocket server (dev)
go run ./cmd/worker             # Start Temporal worker (prod)
go run ./cmd/server             # Start frontend/server (prod)
go run ./cmd/ws                  # Start WebSocket server (prod)
go test ./...                   # Run tests
golangci-lint run               # Lint
```

## Key constraints

- **Workflows are deterministic** — never call time.Now(), random, or external services directly in workflow code
- **Activities are non-deterministic** — put all side effects (DB, HTTP, file I/O) in activities
- **Use signals** for frontend → workflow communication (pause, cancel, input)
- **Use queries** for frontend → workflow state reads (progress %, current step)
- Require `TEMPORAL_HOST_URL` and `TEMPORAL_NAMESPACE` env vars
- Requires Temporal server running locally (see README.md)