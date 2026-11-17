# Backend Agent Instructions

These instructions apply to Go backend development, primarily in `internal/server/`. Follow the root `AGENTS.md` as well.

Inspect relevant source code and tests before making changes. Treat this document as guidance, not as a substitute for verifying current implementation behavior.

## Architecture

- `server.go` handles dependency initialization and wiring.
- `handler/` handles HTTP routes, authentication, request parsing, and responses.
- `app/` contains business logic and service interfaces.
- `repo/` handles persistence and database queries.
- `moex/` handles MOEX integration.
- `internal/model/` contains shared domain and API models.
- `servererrors/` contains domain errors.

Respect existing layer boundaries and interfaces. When changing dependencies, routes, or API contracts, inspect the corresponding initialization, handlers, services, and models.

## Authentication and Authorization

- Public routes include login, user creation, and token refresh.
- Protected routes use `accessTokenValidation` to populate user identity in the request context.
- Propagate authenticated identity through the application and repository layers. Never trust client-supplied ownership identifiers.
- Preserve profile ownership and isolation when changing database operations.

## Portfolio and MOEX Behavior

- Most stock and group operations currently use the user's default profile. Verify profile selection and ownership behavior before modifying them.
- Adding an existing stock merges quantities within its current group; adding it to another group is rejected.
- Stock data is enriched with MOEX pricing and lot information.
- MOEX client results are consumed positionally. Preserve ordering and per-item error handling.
- Group weight updates validate submitted values and their batch total, not necessarily the resulting total across stored groups. Rebalancing performs its own validation.
- Preserve existing rebalancing semantics, including the no-sell option, unless explicitly instructed otherwise.

## Consistency and Transactions

- Some application operations execute multiple repository calls, each potentially using its own transaction. Do not assume the complete operation is atomic.
- When implementing multi-step operations, consider partial failures, transaction boundaries, and data consistency.
- Do not claim that multiple API requests constitute an atomic save without implementing and verifying that guarantee.

## Configuration and Testing

- Inspect `server/config.go` and dependency initialization before changing runtime configuration.
- Be aware that the current development `.env` loader can overwrite existing environment variables.
- Use existing interfaces and generated mocks for isolated tests.
- Follow root `AGENTS.md` for verification commands, mock generation, and Swagger regeneration.
- Add regression tests when modifying business rules, authorization, or persistence behavior.
