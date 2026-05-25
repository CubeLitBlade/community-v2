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
    │   ├── events/            ← RabbitMQ pub/sub (Consumer, Publisher)
    │   └── platform/          ← fx wiring, config, DB, server, logger
    └── services/
        ├── account/           ← user registration, auth, login
        ├── authz/             ← OpenFGA authorization + event-driven tuple writer
        └── post/              ← (skeleton, not yet implemented)
```

### Workspace modules (6)

| Module | Path | Purpose |
|---|---|---|
| `pkg/common` | `backend/pkg/common/` | httperr, idgen, jwt |
| `pkg/events` | `backend/pkg/events/` | RabbitMQ Consumer + Publisher |
| `pkg/platform` | `backend/pkg/platform/` | fx, config, DB, server, logger |
| `services/account` | `backend/services/account/` | Account service |
| `services/authz` | `backend/services/authz/` | Authz service |
| `services/post` | `backend/services/post/` | Post service (stub) |

### Service internal layout

Each service follows the same pattern:

```
services/<name>/
├── cmd/<name>/main.go           ← entrypoint
├── .env                         ← service-specific env vars
├── .envrc                       ← direnv: source_up + dotenv_if_exists .env
└── internal/
    ├── bootstrap/               ← fx app, config, routes
    ├── health/                  ← liveness + readiness handlers
    │   └── setup/module.go
    └── <domain>/
        ├── <domain>.go          ← core domain service
        ├── errors.go            ← sentinel errors
        ├── setup/module.go      ← fx DI module
        └── transport/           ← Gin HTTP handlers + error→HTTP mapping
```

## Commands

All `go` commands run from `backend/` (the workspace root). For service-specific operations, `cd` into the service directory.

```bash
# Build a specific service
cd backend/services/authz && go build ./cmd/authz/

# Run all tests across the workspace
cd backend && go test ./...

# Run a single package's tests
cd backend/services/authz && go test ./internal/authz/...

# Lint a specific service or package
cd backend && golangci-lint run ./services/authz/...

# Lint everything
cd backend && golangci-lint run ./...

# Format (gofumpt + gci)
cd backend/services/authz && gofumpt -l -w . && gci write .
```

### Unpublished pkg modules — `replace` workaround

When a `pkg/` module hasn't been tagged and pushed yet, services that depend on it need a `replace` directive to resolve it locally. For example, authz depends on `pkg/events`:

```bash
cd backend/services/authz
go mod edit -replace=github.com/cubelitblade/community-v2/backend/pkg/events=../../pkg/events
go mod tidy
```

**Before committing**: remove the `replace` directive and bump the version in `require` once the pkg module is published. After publishing, `go mod tidy` resolves normally without the workaround.

## Stack

- **HTTP**: Gin (`gin-gonic/gin`) — `gin.New()` with custom middleware
- **DI**: `go.uber.org/fx` — all services use fx for lifecycle and dependency wiring
- **ORM**: GORM v2 with PostgreSQL driver — `TranslateError: true`
- **Auth**: JWT HS256 via `golang-jwt/jwt/v5`, access token in cookie
- **IDs**: Custom Snowflake generator (`pkg/common/idgen/`)
- **Authorization**: OpenFGA via `openfga/go-sdk` (client on port 9090)
- **Messaging**: RabbitMQ via `rabbitmq/amqp091-go` (port 5672), topic exchange `domain.events`

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
| `RABBITMQ_URL` | AMQP DSN |
| `JWT_SECRET` | Min 32 bytes |

### Authz service vars

| Variable | Default | Notes |
|---|---|---|
| `HTTP_ADDR` | `:8081` | Listen address |
| `FGA_API_URL` | (required) | OpenFGA HTTP API |
| `FGA_STORE_ID` | (required) | OpenFGA store ID |
| `FGA_MODEL_ID` | — | Auth model ID (optional) |

### Account service vars

| Variable | Default | Notes |
|---|---|---|
| `HTTP_ADDR` | `:8080` | Listen address |
| `DATABASE_URL` | (required) | PostgreSQL DSN |
| `SNOWFLAKE_ID` | (required) | Worker ID for Snowflake |
| `ACCESS_TOKEN_COOKIE_NAME` | `access_token` | |
| `COOKIE_SECURE` | `false` | |
| `COOKIE_SAME_SITE` | `lax` | |

## Architecture conventions

- **Unexported struct fields** by default. Only config/params bags (`Config`, `Deps`, `TupleKey`, `Snapshot`) use exported fields. Domain objects and service structs use unexported fields populated via constructor injection.
- **Module Setup pattern**: `internal/<domain>/setup/module.go` provides an fx module. Handlers implement `platform.HTTPMounter` and are annotated with `fx.ResultTags("group:\"mounter\"")` for automatic route registration.
- **Error-to-HTTP mapping**: `httperr.ErrorMapper` funcs in each `transport/` package map domain sentinel errors to RFC 9457 problem details via `httperr.WriteMappedError`.
- **GORM models** are private (`postgres.Row`), mapped via `accountToRow`/`rowToAccount`.
- **Reader/Writer separation**: The `postgres` package exports `Reader` and `Writer`, never exposing GORM directly.
- **The `bootstrap` package** is each service's composition root — it owns config loading, fx app construction, and route registration.
- **Events**: RabbitMQ uses topic exchange `domain.events`. Authz's consumer listens for `user.registered`, `user.role_assigned`, `user.role_revoked`, `account.created`, `post.created` and writes corresponding tuples to OpenFGA.

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
