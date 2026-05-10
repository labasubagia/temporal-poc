# Agent Instructions

This document defines the operational boundaries, project structure, and technical standards for AI agents. These rules are mandatory to minimize hallucinations and maintain the integrity of the Temporal state machine.

---

## 1. Project Context & Tech Stack
- **Primary Language:** Go (Golang)
- **Orchestration:** Temporal (Workflows & Activities)
- **Hot Reload:** `air` for development
- **Linter:** `golangci-lint`

## 2. Codebase Map
- `/cmd`: Entry points for workers and servers.
- `/workflows`: Deterministic business logic (Temporal).
- `/activities`: Side-effect heavy code (DB, HTTP, Disk I/O).
- `/docs`: Detailed design and architecture documentation.

---

## 3. Grounding & Verification Protocol
**To ensure you have read these docs, you must follow this protocol for every task:**

- **Read-Verify-Write:** You are prohibited from assuming the content of any file. Before proposing a change, you MUST use `ls`, `grep`, or `cat` to confirm the current state of the code and documentation.
- **Context Acknowledgment:** Start your first response in a task by stating: *"I have read AGENTS.md and [List other relevant files]."*
- **Traceability:** When suggesting a change, specify exact line numbers or function signatures discovered during your "Read" phase.
- **No Path Guessing:** If a file or directory is missing, do not invent one. Ask the user for the correct path.

---

## 4. Guardrails

### Decision Making
- **When in doubt, always ask** — never assume user intent, preferences, or requirements.
- If a request is ambiguous, stop and request clarification before editing.

### Commit Safety
- **NEVER commit without explicit user permission** — always ask "commit?" first.
- You may stage changes with `git add`, but do NOT execute `git commit` unless explicitly instructed.

### File Scope & Integrity
- **Atomic Changes:** Only modify files directly related to the user's request.
- **No Placeholders:** Never use `// ...` or `// implementation here`. You must provide full, functional code blocks or precise diffs.
- **Dependency Control:** Do not add new Go modules or external packages to `go.mod` without explicit approval.

### Documentation
- **Automatically update relevant docs** when code changes:
  - `README.md` — update for new workflows, routes, or setup changes.
  - `docs/design.md` — update for architectural or workflow changes.
  - `AGENTS.md` — update when adding new constraints or commands.

---

## 5. Technical Constraints (Go & Temporal)

### Determinism Rules
- **Workflows are deterministic:** Never call `time.Now()`, `rand`, or external services (DB/HTTP) directly in workflow code.
- **Activities are non-deterministic:** All side effects (DB, HTTP, file I/O) must reside in activities.
- **Context Check:** Before modifying a function, identify if it is a **Workflow** or an **Activity** and apply the corresponding rules.

### Communication Patterns
- **Use Signals** for frontend → workflow communication (e.g., pause, cancel, input).
- **Use Queries** for frontend → workflow state reads (e.g., progress %, current status).

---

## 6. Code Quality & Style

- **Prefer Guard Clauses:** Use the `if err != nil { return err }` pattern to keep logic flat. Avoid deep `if/else` nesting.
- **Execution-Linked Truth:** After any code change, you must verify the logic by suggesting or running `go build ./...` or `go test ./...`.
- **Linting Gate:** Every task is incomplete until `golangci-lint run` returns zero errors.

---

## 7. Quick Commands

### Development
```sh
air -c air-worker.toml          # Start Temporal worker (dev)
air -c air-server.toml          # Start frontend/server (dev)
air -c air-ws.toml              # Start WebSocket server (dev)
```

### Production / CLI
```sh
go run ./cmd/worker             # Start Temporal worker (prod)
go run ./cmd/server             # Start frontend/server (prod)
go run ./cmd/ws                 # Start WebSocket server (prod)
```

### Quality Control
```sh
go test ./...                   # Run tests
golangci-lint run               # Run linter
go build ./...                  # Verify compilation
```