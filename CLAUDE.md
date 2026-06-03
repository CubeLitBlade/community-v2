# CLAUDE.md

This file provides guidance to Claude Code (claude.ai/code) when working with code in this repository.

## Project layout

Monorepo with a Go workspace at `backend/`. The repo root holds non-Go config.

```
.
├── .golangci.yml              ← golangci-lint v2 config (70+ linters)
├── .env.example               ← template for .env
├── .envrc                     ← direnv: loads .env
├── compose.yaml               ← Docker Compose: db, rabbitmq, openfga
├── compose.override.yaml      ← Dev overrides: port mappings
└── backend/
    ├── go.work                ← Go workspace root
    ├── go.work.sum
    ├── infra/db/init.sql      ← PostgreSQL init: users, schemas, grants
    ├── pkg/
    │   ├── common/            ← shared primitives (httperr, idgen, jwt)
    │   ├── events/            ← outbox pattern, RabbitMQ pub/sub
    │   └── platform/          ← DB, Gin engine, HTTP mounter, middleware
    └── services/
        ├── account/           ← user registration, auth, login
        ├── authz/             ← (skeleton, not yet rebuilt)
        └── post/              ← (skeleton, not yet implemented)
```

### Workspace modules (5)

| Module | Path | Purpose |
|---|---|---|
| `pkg/common` | `backend/pkg/common/` | httperr, idgen, jwt |
| `pkg/events` | `backend/pkg/events/` | Outbox relay, RabbitMQ Consumer + Publisher |
| `pkg/platform` | `backend/pkg/platform/` | DB, Gin engine, HTTP mounter, middleware |
| `services/account` | `backend/services/account/` | Account service |
| `services/post` | `backend/services/post/` | Post service (stub) |

### pkg/platform layout

The `platform` package was flattened from sub-packages into top-level files:

```
pkg/platform/
├── database.go          ← GORM PostgreSQL connection (was database/database.go)
├── gin.go               ← Gin engine setup + RequestIDMiddleware (was server/gin.go)
└── router.go            ← HTTPMounter interface + RegisterRouters helper
```

### Service internal layout

The account service follows this structure:

```
services/account/
├── cmd/account/main.go           ← entrypoint
├── .env                         ← service-specific env vars
├── .envrc                       ← direnv: source_up + dotenv_if_exists .env
├── api/                         ← cross-service contract types
│   ├── events/v1/               ← Event payloads + topic constants
│   └── types/v1/                ← Request/response DTOs
└── internal/
    ├── bootstrap/               ← fx app, config loading, module wiring
    │   ├── app.go               ← fx app construction
    │   ├── config.go            ← top-level Config aggregator
    │   ├── cookie.go            ← cookie security policy (derived from APP_ENV)
    │   ├── database.go          ← GORM DB provider
    │   ├── gin.go               ← Gin engine provider
    │   ├── jwt.go               ← JWT signer + parser providers
    │   ├── logger.go            ← slog.Logger provider
    │   ├── rabbitmq.go          ← RabbitMQ connection + publisher providers
    │   ├── routes.go            ← route registration (API + health)
    │   ├── runtime.go           ← AppEnv + AppRuntime interface
    │   ├── server.go            ← HTTP server provider
    │   ├── snowflake.go         ← Snowflake ID generator provider
    │   └── token.go             ← Token issuer + TTL provider
    ├── contracts/               ← shared types and interfaces (AccessTokenClaims, TTLProvider)
    ├── health/                  ← liveness + readiness handlers
    │   └── setup/module.go
    └── <domain>/
        ├── <domain>.go          ← core domain service
        ├── errors.go            ← sentinel errors
        ├── setup/module.go      ← fx DI module
        └── transport/           ← Gin HTTP handlers + error→HTTP mapping
```

The authz and post services are skeletons — they have go.mod files but minimal or no Go source.

## Commands

All `go` commands run from `backend/` (the workspace root). For service-specific operations, `cd` into the service directory.

```bash
# Build the account service
cd backend/services/account && go build ./cmd/account/

# Run all tests across the workspace
cd backend && go test ./...

# Run a single package's tests
cd backend/services/account && go test ./internal/account/...

# Lint a specific service or package
cd backend && golangci-lint run ./services/account/...

# Lint everything
cd backend && golangci-lint run ./...

# Format (gofumpt + gci)
cd backend/services/account && gofumpt -l -w . && gci write .
```

### Unpublished pkg modules — `replace` workaround

When a `pkg/` module hasn't been tagged and pushed yet, services that depend on it need a `replace` directive to resolve it locally. For example, account depends on `pkg/events`:

```bash
cd backend/services/account
go mod edit -replace=github.com/cubelitblade/community-v2/backend/pkg/events=../../pkg/events
go mod tidy
```

**Before committing**: remove the `replace` directive and bump the version in `require` once the pkg module is published. After publishing, `go mod tidy` resolves normally without the workaround.

## Stack

- **HTTP**: Gin (`gin-gonic/gin`) — `gin.New()` with custom middleware
- **DI**: `go.uber.org/fx` — all services use fx for lifecycle and dependency wiring
- **ORM**: GORM v2 with PostgreSQL driver — `TranslateError: true`, generic `gorm.G[T]` API
- **Auth**: JWT HS256 via `golang-jwt/jwt/v5`, access token in cookie
- **IDs**: Custom Snowflake generator (`pkg/common/idgen/`)
- **Authorization**: OpenFGA via `openfga/go-sdk` (client on port 9090)
- **Messaging**: RabbitMQ via `rabbitmq/rabbitmq-amqp-go-client` (port 5672), topic exchange `domain.events`
- **Events**: CloudEvents spec via `cloudevents/sdk-go/v2`

## Infrastructure (Docker Compose)

| Service | Port(s) | Notes |
|---|---|---|
| PostgreSQL 18 | 5432 | Users: `goose`, `account_service_svc`, `openfga` |
| RabbitMQ 4.3 | 5672, 15672 (mgmt) | `community` / `community_dev_password` |
| OpenFGA | 9090 (HTTP), 9091 (gRPC) | Playground disabled |

## Config

All config via **environment variables**. Use `direnv allow` to auto-load `.env` via `.envrc`. Root `.env` holds shared vars; each service has its own `.env` with service-specific vars.

### Shared vars (root `.env`)

| Variable | Notes |
|---|---|
| `APP_ENV` | `dev` or `prod` |
| `JWT_SECRET` | Min 32 bytes |

### Account service vars

| Variable | Default | Notes |
|---|---|---|
| `APP_ENV` | `dev` | `dev` or `prod` |
| `HTTP_ADDR` | `:8080` | Listen address |
| `DATABASE_URL` | (required) | PostgreSQL DSN |
| `SNOWFLAKE_ID` | (required) | Worker ID for Snowflake |
| `RMQ_URL` | (required) | RabbitMQ AMQP DSN |
| `RMQ_SYSTEM_EXCHANGE_NAME` | (required) | RabbitMQ exchange name |
| `JWT_SECRET` | (required) | Min 32 bytes |
| `JWT_ISSUER` | `community-v2` | Token issuer claim |
| `ACCESS_TOKEN_TTL` | `2h` | Access token validity duration |

Cookie security (Secure/SameSite) is derived automatically from `APP_ENV`: `dev` → insecure/lax, `prod` → secure/lax.

## Architecture conventions

- **Unexported struct fields** by default. Only config/params bags (`Config`, `Deps`, `TupleKey`, `Snapshot`) use exported fields. Domain objects and service structs use unexported fields populated via constructor injection.
- **Module Setup pattern**: `internal/<domain>/setup/module.go` provides an fx module. Handlers implement `platform.HTTPMounter` and are annotated with `fx.ResultTags("group:\"mounter\"")` for automatic route registration via `platform.RegisterRouters`.
- **Error-to-HTTP mapping**: `httperr.ErrorMapper` funcs in each `transport/` package map domain sentinel errors to RFC 9457 problem details via `httperr.WriteMappedError`.
- **GORM models** are private (`storage.EventRow`, `storage.AccountRow`), mapped via constructor functions.
- **Reader/Writer separation**: The `storage` package exports separate reader and writer types, never exposing GORM directly.
- **The `bootstrap` package** is the service's composition root — split into focused files (config, database, gin, jwt, logger, rabbitmq, runtime, server, snowflake, token, cookie) plus `app.go` for fx app construction and `routes.go` for route registration.
- **Contracts package**: `internal/contracts/` holds shared types and interfaces (e.g. `AccessTokenClaims`, `TTLProvider`) used across domain boundaries within the service.
- **API package**: `api/` holds cross-service contract types (event payloads, request/response DTOs) organized by version (`v1`).
- **Events**: Outbox relay pattern in `pkg/events/outbox/` polls the database for unpublished events, wraps them as CloudEvents, publishes to RabbitMQ, and marks them as published.

## Testing

- Tests use `t.Parallel()` everywhere. Use table-driven tests with helper assertion funcs.
- No integration tests; no test DB fixtures.
- White-box tests (same package) use `//nolint:testpackage` directive.
- Stub types implementing third-party interfaces need `//nolint:ireturn` and `//nolint:revive` annotations for SDK method names.

## Lint strictness

The `.golangci.yml` is aggressive (70+ linters). Key pain points:
- **depguard**: Only allows stdlib + listed third-party imports. Add new dependencies to `.golangci.yml` `linters.settings.depguard.rules.main.allow`.
- **exhaustruct**: All structs must be fully initialized. Use `//nolint:exhaustruct` only for sentinel zero-value returns where callers check the bool.
- **revive** `exported`: Every exported symbol needs a doc comment.
- **wrapcheck**: Errors from external packages and interface methods must be wrapped.
- **varnamelen**: Short names only for specific patterns (`c *gin.Context`, `db *gorm.DB`, `tt *testing.T`, etc.).
- **funlen** (80 lines) / **cyclop** (max 10): strict complexity limits.
- **tagliatelle**: JSON tags must be camelCase. Add `//nolint:tagliatelle` on payload structs whose JSON keys are defined by cross-service contract.
- **nolintlint**: All `//nolint` directives must include an explanation.
