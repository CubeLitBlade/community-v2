# CLAUDE.md

This file provides guidance to Claude Code (claude.ai/code) when working with code in this repository.

## Project layout

Single Go module — the module root is `backend/`, **not** the repo root.
The repo root holds non-Go config: `.golangci.yml`, `.env.example`, `qodana.yaml`.

```
.
├── backend/
│   ├── go.mod              ← module github.com/CubeLitBlade/community-v2/backend
│   ├── cmd/api/main.go     ← single entrypoint
│   ├── db/schema.sql       ← PostgreSQL schema (accounts table)
│   └── internal/
│       ├── bootstrap/      ← app wiring, config, DB, router, server
│       ├── account/        ← domain model + service layer
│       │   ├── setup/      ← DI wiring & route registration
│       │   ├── transport/  ← HTTP handlers + error→HTTP mapping
│       │   └── postgres/   ← GORM repository (Reader/Writer)
│       ├── auth/           ← JWT issuer + login service
│       │   ├── setup/
│       │   └── transport/
│       ├── idgen/          ← custom Snowflake ID generator
│       └── httperr/        ← RFC 9457 Problem Details helpers
```

## Commands

All `go` commands **must** run from `backend/`:

```bash
# Run all tests
cd backend && go test ./...

# Run a single test
cd backend && go test -run TestRegister ./internal/account/

# Lint (golangci-lint v2 config at repo root)
cd backend && golangci-lint run ./...

# Format (gofumpt + gci)
cd backend && gofumpt -l -w . && gci write .
```

## Stack

- **HTTP**: Gin (`gin-gonic/gin`) — uses `gin.New()` (no default middleware), with custom `gin.Logger()` + `gin.Recovery()`
- **ORM**: GORM v2 with PostgreSQL driver — `TranslateError: true` enabled
- **Auth**: JWT HS256 via `golang-jwt/jwt/v5`, access token in cookie
- **IDs**: Custom Snowflake generator (`internal/idgen/`) — epoch 2026-03-11

## Config

All config via **environment variables** (no config files). Key vars:

| Variable | Default | Notes |
|---|---|---|
| `ADDR` | `:8080` | Listen address |
| `DATABASE_URL` | (required) | PostgreSQL DSN |
| `JWT_SECRET` | (required) | Min 32 bytes |
| `SNOWFLAKE_ID` | `1` | Worker ID for Snowflake |
| `JWT_ACCESS_TOKEN_TTL` | `2h` | Duration string |
| `ACCESS_TOKEN_COOKIE_NAME` | `ACCESS_TOKEN` | |
| `COOKIE_SECURE` | `false` | |

Use `direnv allow` to auto-load `.env` via `.envrc`.

## Architecture conventions

- **All struct fields are unexported** unless the struct is a config/params bag (`Config`, `ModuleDeps`, `Deps`, `Snapshot`). Domain objects and service structs (`Registrar`, `Authenticator`, `Login`, `JWTIssuer`) use unexported fields populated via constructor injection.
- **Module Setup pattern**: `internal/<domain>/setup/module.go` wires dependencies and calls `handler.RegisterRoutes()` on the Gin router. Each module exposes a `Mount(router gin.IRouter)` method.
- **Error-to-HTTP mapping**: `httperr.ErrorMapper` funcs declared in each `transport/` package. Domain sentinel errors are mapped 1:1 to RFC 9457 problem details via `httperr.WriteMappedError`.
- **GORM models** are private (`postgres.Row`), mapped to domain objects via `accountToRow` / `rowToAccount`.
- **Reader/Writer separation**: The `postgres` package exports `Reader` (queries) and `Writer` (mutations), never exposing GORM directly.
- **The `bootstrap` package** is the composition root — it owns `OpenDatabase`, `LoadConfig`, `newRouter`, `RegisterModules`.

## Testing

- Tests use `t.Parallel()` everywhere. Use table-driven tests with helper assertion funcs.
- No integration tests; no test DB fixtures.
- `internal/account` tests use `//nolint:testpackage` (white-box tests in `package account`), others use `_test` suffix (black-box).

## Lint strictness

The `.golangci.yml` is aggressive (70+ linters). Key pain points:
- **depguard**: Only allows stdlib + listed third-party imports. Add new dependencies to `.golangci.yml` `linters.settings.depguard.rules.main.allow`.
- **exhaustruct**: All structs in `internal/` must be fully initialized.
- **revive** `exported`: Every exported symbol needs a doc comment.
- **wrapcheck**: Errors from external packages and interface methods must be wrapped.
- **varnamelen**: Short names only allowed for specific patterns (`c *gin.Context`, `db *gorm.DB`, `tt *testing.T`, etc.).
- **funlen** (80 lines) / **cyclop** (max 10) / **nestif** (min 4): strict complexity limits.
- GORM error signatures (`.Error`, `.Find()`, `.Create()`, etc.) and stdlib patterns (`.Close()`, `.Decode()`, etc.) are excluded from many checks via `ignoreSigs`.
