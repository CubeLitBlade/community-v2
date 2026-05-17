# AGENTS.md

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

- **HTTP**: Gin (`gin-gonic/gin`) — uses `gin.New()` (no default middleware), with custom `gin.Logger()` + `gin.Recovery()` + Cross-Origin protection
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

- Domain objects (`account.Account`) are **unexported struct fields** — access via getter methods and `Snapshot()`. Reconstruction from DB uses `NewAccountFromSnapshot`.
- Modules follow a Setup pattern: `internal/<domain>/setup/module.go` wires dependencies and calls `handler.RegisterRoutes()` on the Gin router.
- Error-to-HTTP mapping uses `httperr.ErrorMapper` funcs declared in each `transport/` package. Domain sentinel errors are mapped 1:1 to RFC 9457 problem details.
- GORM models are private (`postgres.Row`), mapped to domain objects via `accountToRow` / `rowToAccount`.
- The `bootstrap` package is the composition root — it owns `OpenDatabase`, `LoadConfig`, `newRouter`, `RegisterModules`.

## Testing

- Only `internal/account` has tests (white-box, `//nolint:testpackage`).
- `internal/account/password_test.go` is entirely **commented out** — do not uncomment blindly; the password validation API changed.
- Tests use `t.Parallel()` everywhere. Use table-driven tests with helper assertion funcs.
- No integration tests; no test DB fixtures.

## Lint strictness

The `.golangci.yml` is aggressive (70+ linters). Notable pain points:
- **depguard**: Only allows stdlib + listed third-party imports. Add new dependencies to `.golangci.yml` `depguard.rules.main.allow`.
- **exhaustruct**: All structs in `internal/` must be fully initialized.
- **revive** `exported`: Every exported symbol needs a doc comment.
- **wrapcheck**: Errors from external packages and interface methods must be wrapped.
- **varnamelen**: Short names only allowed for specific patterns (`c *gin.Context`, `db *gorm.DB`, `tt *testing.T`, etc.).
- **funlen** (80 lines) / **cyclop** (max 10) / **nestif** (min 4): strict complexity limits.
- GORM error signatures (`.Error`, `.Find()`, `.Create()`, etc.) and stdlib patterns (`.Close()`, `.Decode()`, etc.) are excluded from many checks via `ignoreSigs`.
