# Agent Instructions

## Project

Portfolio management application that calculates buy/sell recommendations to restore target allocations.

**Stack:** Go HTTP backend and CLI, PostgreSQL, JWT authentication, MOEX integration. SvelteKit + TypeScript frontend planned.

Read `README.md` for project context, but treat the code as the source of truth. Ignore `notes.md`. Do not confuse planned features with implemented functionality.

Before working on any frontend-related code, read and follow the instructions in `docs/frontend.md`.

## Repository Structure

- `cmd/server/`, `cmd/cli/` — entrypoints
- `internal/server/handler/` — HTTP layer
- `internal/server/app/` — business logic
- `internal/server/repo/` — database access
- `internal/server/moex/` — MOEX integration
- `internal/model/` — shared models
- `internal/cli/` — CLI implementation
- `pkg/tokenmanager/` — JWT handling
- `migrations/` — SQL migrations
- `docs/`, `mocks/` — generated artifacts

## Working Rules

- Inspect relevant code and tests before making changes.
- Make focused changes; avoid unrelated refactoring or functionality.
- Mark temporary code with a nearby `TEMP` comment.
- Follow established project conventions and architecture.
- Ask before making consequential, unspecified architectural or product decisions.
- Preserve unrelated worktree changes.
- Report changes, verification performed, and remaining limitations.
- Never read or expose credentials, tokens, or sensitive configuration unnecessarily.

## Architecture

- Keep HTTP handling, business logic, persistence, and MOEX integration in their respective layers.
- Business logic, including financial calculations, belongs in the Go backend.
- Preserve interface boundaries for dependency injection and testing.
- Enforce user ownership and profile isolation on the backend.
- When changing API contracts, check compatibility with existing clients, particularly the CLI.

## Domain Rules

Unless explicitly changed by the task:

- Holdings quantities represent integer lots.
- Holding value = price × lot size × lot count.
- Rebalancing requires group weights totaling 100%.
- Holdings within a group receive equal monetary allocations.
- Rebalancing is approximate due to integer-lot constraints.
- Rebalancing generates proposals; it never executes broker orders.
- Existing operations generally use the user's default profile.

## Important Constraints

- Repository scanning in `internal/server/repo/helper.go` relies on struct field declaration order. Check associated SQL queries when changing model fields.
- Use new numbered up/down migrations for schema changes; don't rewrite applied migrations or reset data without approval.
- Modify source annotations/interfaces before regenerating Swagger documentation or mocks.
- Do not manually edit generated artifacts.

## Verification

For Go changes, run appropriate checks:

```bash
go test ./...
go vet ./...
go build ./...
```

Add regression tests when changing behavior. Use mocked MOEX responses for deterministic tests.

Generate mocks with **Mockery** (`mockery`) and Swagger documentation with **Swaggo** (`swag`). Modify source interfaces and annotations first; never edit generated files manually.
- Regenerate mocks: `mockery`
- Regenerate Swagger docs: `swag init -g internal/server/handler/swagger_debug.go -o docs`

Use the project's existing configuration and adjust commands to match its actual setup.

Do not claim tests or commands passed unless they actually ran. Avoid stateful operations, such as starting Docker Compose or resetting databases, without considering their effects.
