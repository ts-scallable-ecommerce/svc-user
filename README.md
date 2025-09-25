# User Service (svc-user)

This repository provides the Go implementation of the User Service described in the scalable ecommerce platform technical documentation. It focuses on identity, authentication, and RBAC responsibilities and is prepared for deployment on Kubernetes behind Traefik without relying on Docker Compose.

## Features

- Fiber HTTP server exposing health checks and ready for REST handlers defined in the OpenAPI specification.
- gRPC server bootstrap for inter-service communication.
- PostgreSQL access layer with migrations aligned to the documented schema.
- Redis client helpers for caching, token revocation, and rate limiting primitives.
- Argon2id password hashing utilities and RSA-based JWT token issuer helpers.
- Kafka event producer suitable for transactional outbox dispatch.
- Modular internal packages covering users, RBAC, and configuration loading.

## Project Layout

```
cmd/
  user-http/      # Fiber HTTP entrypoint
  user-grpc/      # gRPC entrypoint
internal/
  auth/           # JWT + password helpers
  cache/          # Redis client helpers
  config/         # Environment configuration loader
  db/             # Database utilities and migrations
  events/         # Kafka producer
  grpc/           # gRPC server helpers
  http/           # HTTP router, handlers, middleware
  rbac/           # Permission resolution helpers
  users/          # User domain repository & service
api/
  openapi.yaml    # User Service HTTP API definition
proto/
  user.proto      # gRPC contracts
```

## Getting Started

### Prerequisites

- Go 1.22+
- PostgreSQL 15+
- Redis 7+
- Kafka 3+

### Environment Variables

| Variable | Description |
| --- | --- |
| `DB_DSN` | PostgreSQL DSN (e.g. `postgres://user:pass@localhost:5432/usersvc?sslmode=disable`) |
| `REDIS_ADDR` | Redis address (default `localhost:6379`) |
| `HTTP_ADDR` | Fiber listen address (default `:8080`) |
| `GRPC_ADDR` | gRPC listen address (default `:9090`) |
| `JWT_PRIVATE_KEY_PATH` | Path to RSA private key for signing |
| `JWT_PUBLIC_KEY_PATH` | Path to RSA public key for verification |

### Commands

```
make run-http    # Run the Fiber HTTP server
make run-grpc    # Run the gRPC server
make build       # Build binaries into ./bin
make test        # Execute Go tests
make migrate-up  # Apply database migrations (requires migrate tool)
make migrate-down# Rollback the last migration
```


The repository includes a multi-stage `Dockerfile` that builds both the HTTP and gRPC binaries using Go 1.25.1 and packages them
into a minimal distroless runtime image.

```bash
docker build -t svc-user:local .
docker run --rm -p 8080:8080 svc-user:local               # Start the HTTP server
docker run --rm svc-user:local /usr/local/bin/user-grpc   # Start the gRPC server
```

Override environment variables (e.g., database credentials, Redis address) with `-e` flags or a mounted file when running the
container.

### Migrations

The SQL migrations under `internal/db/migrations` implement the schema defined in the technical specification, including RBAC tables and a transactional outbox table. Use the [`golang-migrate`](https://github.com/golang-migrate/migrate) CLI or compatible tooling to apply them.

### OpenAPI & Protobuf Contracts

- `api/openapi.yaml` mirrors the documented REST API, suitable for generating client SDKs or validating handlers.
- `proto/user.proto` defines the gRPC interfaces consumed by other services. Use [`buf`](https://buf.build) for linting and generating code in the multi-repo strategy.

## Testing

Unit tests can be executed via `make test`. Integration tests should be configured to spin up PostgreSQL, Redis, and Kafka containers using your preferred tooling (e.g., Tilt, Skaffold, or manual container orchestration). Docker Compose is intentionally omitted to align with the requirements.

## Observability & Operations

The service is ready for OpenTelemetry instrumentation, Prometheus metrics, and structured logging. Kubernetes manifests and Traefik configuration snippets referenced in the documentation should live in their respective repositories within the multi-repo strategy.

