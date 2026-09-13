# Task Manager

A small backend microservice for managing to-do tasks, built with Go, Gin, and PostgreSQL. Built as part of the Golang Developer hiring assessment for Graph.

## Tech Stack

- **HTTP framework:** [Gin](https://github.com/gin-gonic/gin) (required by the assessment)
- **Database:** PostgreSQL, accessed via [sqlx](https://github.com/jmoiron/sqlx) + the [pgx/v5](https://github.com/jackc/pgx) stdlib driver
- **Testing:** Go standard library `testing` package only (no third-party assertion or mocking libraries) — see [Design Decisions](#design-decisions--trade-offs)
- **Docs:** OpenAPI/Swagger via [swaggo](https://github.com/swaggo/swag)
- **Observability:** Prometheus metrics + request-ID based log correlation
- **Containerization:** Docker (multi-stage build) + Docker Compose

## Project Structure

```
task-manager/
├── cmd/api/
├── internal/
│   ├── handler/
│   ├── service/
│   ├── repository/
│   ├── model/
│   ├── dto/
│   ├── metrics/
│   └── tracing/
├── migrations/
├── docs/
├── Dockerfile
├── docker-compose.yml
└── .env.example
```

Request flow: **Handler → Service → Repository**. Handlers only decode/encode HTTP and do shallow validation; the Service layer holds business logic, default values, and query validation; the Repository layer only runs SQL.

## Running the Project

### Prerequisites

- Docker and Docker Compose

### Steps

```bash
cp .env.example .env      # adjust values if needed
docker compose up --build -d
```

This starts two containers:

- `postgres` — PostgreSQL 16, with the schema in `migrations/` applied automatically on first boot
- `app` — the API server, listening on `:8080`

Check it's up:

```bash
curl localhost:8080/health
```

### Environment Variables

| Variable       | Description                     | Example (docker-compose)                                                 |
| -------------- | ------------------------------- | ------------------------------------------------------------------------ |
| `DATABASE_URL` | Postgres connection string      | `postgres://taskuser:taskpass@postgres:5432/taskmanager?sslmode=disable` |
| `PORT`         | HTTP port the server listens on | `8080`                                                                   |

When running the Go binary directly on the host (not via Docker), `DATABASE_URL` should point to `localhost` instead of `postgres` (see `.env.example`).

## API Documentation

Interactive Swagger UI, available once the service is running:

```
http://localhost:8080/swagger/index.html
```

The raw spec is also committed at `docs/swagger.json`.

### Endpoints

| Method | Path         | Description                                  |
| ------ | ------------ | -------------------------------------------- |
| GET    | `/health`    | Liveness check                               |
| GET    | `/metrics`   | Prometheus metrics                           |
| POST   | `/tasks`     | Create a task                                |
| GET    | `/tasks`     | List tasks (supports filtering & pagination) |
| GET    | `/tasks/:id` | Get a task by ID                             |
| PUT    | `/tasks/:id` | Partially update a task                      |
| DELETE | `/tasks/:id` | Delete a task                                |

### curl Examples

**Create a task:**

```bash
curl -X POST localhost:8080/tasks \
  -H "Content-Type: application/json" \
  -d '{
    "title": "Learn Go",
    "description": "Document the API",
    "assignee": "amir",
    "priority": "high"
  }'
```

Response (`201 Created`):

```json
{
  "id": 1,
  "title": "Learn Go",
  "description": "Document the API",
  "status": "pending",
  "priority": "high",
  "assignee": "amir",
  "created_at": "2026-09-12T10:00:00Z",
  "updated_at": "2026-09-12T10:00:00Z"
}
```

**List tasks:**

```bash
curl "localhost:8080/tasks"
```

Response (`200 OK`):

```json
[
  {
    "id": 1,
    "title": "Learn Go",
    "status": "pending",
    "priority": "high",
    "assignee": "amir",
    "...": "..."
  }
]
```

**Get a task:**

```bash
curl localhost:8080/tasks/1
```

**Update a task (partial):**

```bash
curl -X PUT localhost:8080/tasks/1 \
  -H "Content-Type: application/json" \
  -d '{"status": "in_progress"}'
```

Only the fields present in the body are changed; omitted fields keep their existing value.

**Delete a task:**

```bash
curl -X DELETE localhost:8080/tasks/1 -i
```

Returns `204 No Content` on success, `404 Not Found` if the ID doesn't exist.

### Request/Response Reference

**`CreateTaskRequest`**

| Field         | Type   | Required | Notes                                                          |
| ------------- | ------ | -------- | -------------------------------------------------------------- |
| `title`       | string | yes      |                                                                |
| `description` | string | no       |                                                                |
| `assignee`    | string | yes      |                                                                |
| `status`      | string | no       | one of `pending`, `in_progress`, `done`; defaults to `pending` |
| `priority`    | string | no       | one of `low`, `medium`, `high`; defaults to `medium`           |

**`UpdateTaskRequest`** — same fields as above, all optional; only provided fields are changed.

**`TaskResponse`**

```json
{
  "id": 1,
  "title": "string",
  "description": "string",
  "status": "pending | in_progress | done",
  "priority": "low | medium | high",
  "assignee": "string",
  "created_at": "RFC3339 timestamp",
  "updated_at": "RFC3339 timestamp"
}
```

## Running Tests

**Unit tests only** (handler, service, model — no external dependencies, fast):

```bash
go test ./internal/handler/... ./internal/service/... ./internal/model/... -v
```

**Integration tests** (repository layer — needs a running Postgres):

```bash
docker compose up -d postgres
DATABASE_URL="postgres://taskuser:taskpass@localhost:5432/taskmanager?sslmode=disable" \
  go test ./internal/repository/... -v
```

Without `DATABASE_URL` set, these tests are skipped rather than failing.

**All tests + coverage:**

```bash
DATABASE_URL="postgres://taskuser:taskpass@localhost:5432/taskmanager?sslmode=disable" \
  go test ./internal/... -coverprofile=coverage.out
go tool cover -func=coverage.out
```

Current total coverage: **~72%** (requirement: ≥70%). Coverage is measured on `internal/...` only; `cmd/api` is thin dependency-wiring code and is excluded, consistent with common Go practice.

## Observability

**Prometheus metrics** at `GET /metrics`:

- `requests_total{method, path, status}` — request counter
- `request_latency_histogram{method, path}` — request duration histogram
- `tasks_count` — current number of tasks (gauge)

**Tracing:** each request is assigned a request ID (either generated, or propagated from an incoming `X-Request-ID` header), returned in the response header and included in log lines across the Handler and Service layers. This gives request correlation across layers without pulling in a full distributed-tracing stack — see trade-offs below.

## Design Decisions & Trade-offs

- **Layered architecture (Handler → Service → Repository) with separate DTO and Model types.** Keeps the API contract (DTO) independent from the database shape (Model), and keeps business logic (defaults, validation, partial-update merging) out of both the HTTP layer and the data-access layer.
- **Standard library `testing` only, no testify/gomock.** Mocks are hand-written structs with function-typed fields. This is more verbose than using a mocking library, but every line of test code is something we can fully explain and defend, with no "magic" from a third-party library.
- **`context.Context` threaded through every layer**, from the incoming HTTP request down to the SQL calls (`ExecContext`/`GetContext`/`QueryRowxContext`). This enables proper cancellation propagation and is also what carries the request ID used for tracing.
- **sqlx + pgx/v5 (stdlib driver) instead of a full ORM.** Gives direct control over SQL (useful for an assessment where SQL competency likely matters) while avoiding the boilerplate of raw `database/sql`.
- **Auto-increment integer IDs (`BIGSERIAL`) instead of UUIDs.** Simpler for this scope; the trade-off is that IDs are sequential/guessable, which would be reconsidered for a public-facing API where ID enumeration is a concern.
- **Status and priority are validated string-based enums** (`TaskStatus`, `TaskPriority` with an `IsValid()` method), enforced both at the application layer and via `CHECK` constraints in Postgres. Both are optional on creation and default to `pending` / `medium` respectively.
- **Plain SQL migration file, no migration framework.** For a single-migration, 4-day scope, an extra dependency like `golang-migrate` wasn't justified; for a longer-lived project, it would be the natural next step.
- **Lightweight tracing via request-ID log correlation, not full OpenTelemetry.** Matches the assessment's "basic observability" requirement without the operational overhead of running a tracing backend (Jaeger/Tempo) alongside the service.
- **Multi-stage Docker build** (`golang:1.25-alpine` builder → `alpine` runtime, `CGO_ENABLED=0` for a static binary) keeps the final image small while the build stage carries the full Go toolchain.

## Possible Future Improvements

Not implemented in this submission due to the 4-day scope, but straightforward extensions of the current architecture:

- Pagination and filtering on `GET /tasks`
- Redis cache-aside for `GET /tasks`, with invalidation on update/delete — would live in the Service layer, which already sits at the right boundary for this
- A short load test / benchmark with a `pprof` report
