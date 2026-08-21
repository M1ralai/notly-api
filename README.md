# Notly API

Notly API is a modular Go backend for notes, tasks, schedules, habits, and personal planning. It exposes a REST API and real-time notification channel backed by PostgreSQL.

## What this project demonstrates

- A modular-monolith layout with HTTP, service, repository, and domain boundaries
- PostgreSQL persistence with versioned SQL migrations
- JWT access tokens and a database-backed refresh-token flow
- WebSocket delivery for real-time notifications
- Google Calendar OAuth and encrypted storage of third-party tokens
- Prometheus metrics and structured logging with Zap
- Optional email, bot-protection, scheduled-job, and S3-compatible attachment integrations

## Architecture

```text
cmd/api
  -> internal/app                 application composition and route registration
  -> internal/modules/<domain>    HTTP -> service -> repository -> domain
  -> internal/infrastructure      database, auth middleware, jobs, logging,
                                  metrics, OAuth, encryption and WebSockets
```

Business areas are separated under `internal/modules`, including authentication, notes, tasks, schedules, habits, calendar integration, notifications, and users. The application layer creates each dependency and registers its routes; modules depend on shared infrastructure through explicit constructors rather than global state.

A typical authenticated request follows this path:

```text
HTTP request -> middleware -> module handler -> service -> PostgreSQL repository
                                                     -> WebSocket/event side effects
```

Database migrations live in `internal/infrastructure/database/migrations` and run during application startup. Google OAuth tokens are passed through the encryption component before persistence.

## Technology

- Go 1.25
- PostgreSQL, `sqlx`, and `golang-migrate`
- `gorilla/mux` and `gorilla/websocket`
- `golang-jwt`
- Prometheus client and Zap
- MinIO/S3-compatible storage for optional attachments

## Run locally

Requirements: Go 1.25+, PostgreSQL, and optionally the `migrate` CLI for the Makefile migration commands.

```bash
git clone https://github.com/M1ralai/notly-api.git
cd notly-api
cp .env.example .env
go mod download
go run cmd/api/main.go
```

Update the database connection and secrets in `.env` before starting the service. Useful Make targets:

```bash
make run
make build
make test
make migrate-up
```

The repository also includes a Docker Compose stack:

```bash
make docker-config
make docker-up
```

## Example

With the API running on the default port:

```bash
curl http://localhost:8080/health
```

Swagger UI is available locally at `http://localhost:8080/swagger/index.html`. Representative routes include `/api/auth/register`, `/api/auth/login`, `/api/auth/refresh`, `/api/notes`, `/api/tasks`, `/api/calendar/status`, `/metrics`, and `/ws`.

## Optional integrations

Google Calendar, Resend email, Turnstile, and MinIO require their corresponding environment variables from `.env.example`. Leaving an integration unconfigured disables or limits that integration; it does not affect the core database-backed API.

## Limitations

- The repository does not yet contain a comprehensive automated test suite for its feature surface.
- OAuth, email, bot protection, and object storage require external services and separate credentials.
- The included Compose and Nginx configuration is a deployment example; production use still requires independent TLS, secret-management, backup, monitoring, and hardening decisions.

## License

MIT
