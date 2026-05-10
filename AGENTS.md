# Agent Instructions

This document provides context and rules for AI agents working on this project. Treat these as hard constraints to minimize hallucinations and ensure system integrity.

## Project Context & Tech Stack
- **Primary Language:** Go (Golang)
- **Orchestration:** Temporal (Workflows & Activities)
- **Hot Reload:** `air` for development
- **Linter:** `golangci-lint`

## Codebase Map
- `/cmd`: Entry points for workers and servers.
- `/workflows`: Deterministic business logic (Temporal).
- `/activities`: Side-effect heavy code (DB, HTTP, Disk I/O).
- `/docs`: Detailed design and architecture documentation.

---

## The Golden Rules (Hallucination Prevention)

1. **Read-Verify-Write Loop:** 
   - Never assume a function's signature or a file's existence based on training data.
   - You MUST run `ls`, `grep`, or `cat` to verify local context before proposing a change.
   - If a search returns no results, ask for the path; do not guess.

2. **No Placeholders:** 
   - Never use `// ... rest of implementation` or `// code goes here`. 
   - Hallucinations often hide in omitted code. Provide full file contents or precise, verifiable diffs.

3. **Deterministic Workflows:** 
   - If modifying `/workflows`, ensure zero non-determinism. 
   - NO `time.Now()`, `rand`, or external calls. These must stay in `/activities`.

4. **Verify Before Completion:** 
   - After writing code, check for syntax errors using `go build ./...` or `go test ./...`.
   - A task is not "done" until it passes `golangci-lint run`.

---

## Guidelines & Constraints

### Decision Making
- **Ask First:** If requirements are ambiguous, stop and ask.
- **Commit Safety:** You may stage changes (`git add`), but you are strictly forbidden from executing `git commit` without an explicit "yes" to a "commit?" prompt.

### File Scope & Style
- **Atomic Changes:** Only touch files directly relevant to the feature.
- **Guard Clauses:** Always favor `if err != nil { return err }` to keep logic flat. Avoid deep nesting.
- **Documentation:** If you change a workflow, API endpoint, or config, you must update `README.md` or `docs/design.md` immediately.

---

## Development Workflow

### Quick Commands
```sh
# Development
air -c air-worker.toml          # Start Temporal worker
air -c air-server.toml          # Start frontend/server
air -c air-ws.toml              # Start WebSocket server

# Testing & Quality
go test ./...                   # Run all tests
golangci-lint run               # Run linter
go build ./...                  # Verify compilation
```
