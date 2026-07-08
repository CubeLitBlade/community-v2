# Backend AI Agent Rules

This repository is a Go monorepo backend system located under `backend/`.

It consists of multiple microservices and shared infrastructure packages.

---

## 1. Architecture Overview

The system follows a **hexagonal architecture direction**.

- Domain logic: `internal/domain`
- Ports (interfaces): `internal/domain/port`
- Application services (use cases): `internal/application`
- Adapters — driven (DB / external systems): `internal/adapter/driven`
- Adapters — driving (gRPC / HTTP): `internal/adapter/driving`
- Bootstrap: fx dependency wiring only — no business logic

All services MUST follow hexagonal structure. The account service has been fully rewritten to DDD/hexagonal and is the canonical reference implementation.

Additionally, the codebase uses an **OpenTelemetry decorator pattern** for cross-cutting observability:

- Telemetry decorators in `internal/telemetry/` wrap application services (use cases) and adapters
- Decorators record traces, metrics, and structured logs without touching domain logic
- Example: `InstrumentedRegistrar` wraps `Registrar` — the domain never imports OTel

---

## 2. Service Communication Model

There are three valid communication mechanisms:

### 1. Synchronous (East-West)
- gRPC is the standard for service-to-service communication
- Legacy HTTP calls may exist but MUST NOT be extended

### 2. External (North-South)
- External HTTP requests go through an API Gateway
- Gateway translates HTTP → internal gRPC calls

### 3. Asynchronous
- Event-driven messaging using RabbitMQ
- Outbox pattern is used for reliable event publishing

---

## 3. Service Boundaries

Each service under `services/*` is an independent bounded context.

- No direct imports between services
- No shared internal business logic
- Communication only via gRPC or events

---

## 4. Contracts (API & Events)

- Each service owns its own API (proto) and gRPC handler definitions
- Event **payloads** (not topics/constants) may live in `pkg/events/<service>/` when shared between publisher and consumer services
- Event topics, types, and aggregate constants are co-located with the payload in pkg as the canonical shared contract
- Cross-service communication relies on versioned schemas

Semantic compatibility is required, structural differences are allowed.

---

## 5. Shared Packages (pkg/)

`pkg/` contains infrastructure-only code.

### Allowed:
- platform (server, DB, routing)
- events (outbox + RabbitMQ + shared event payloads)
- common (IDs, JWT, errors)

### Rules:
- No business logic in pkg/
- No domain-specific abstractions in pkg/
- Prefer duplication over incorrect abstraction

---

## 6. Shared Package Versioning & Release Workflow

When proto definitions or pkg/ shared packages are added or modified, they MUST follow this strict two-phase
commit and release workflow to maintain dependency integrity:

### Phase 1: Local Integration (The replace commit)

1. Update the consumer service's go.mod to use a replace directive pointing to the local modified pkg/ Or proto path.
2. Commit this state with the replace directive active.

### Phase 2: Versioning & Publishing (The tagged release)

1. Determine the new semantic version:
    - Attempt to use `gorelease` to infer the version bump automatically.
    - Fallback: If `gorelease` fails, times out, or produces no result, MUST infer the correct version bump based on Go compatibility rules
      (e.g., vo.x.o for backwards-compatible changes, v1.0.0 for breaking changes).
1. Create a Git tag for the specific pkg/ or proto module with the new version.
2. Push the tag to the remote repository.
3. Update the consumer service's go.mod to remove the replace directive and use the actual, newly versioned
module path.
1. Commit this final state.

---

## 7. Persistence Strategy (Learning Context)

This repository intentionally uses different persistence approaches:

- Some services use GORM
- Some services use ent

Each service MUST use only one persistence approach internally.

---

## 8. Forbidden Patterns

- No new service-to-service HTTP calls
- No shared domain models across services
- No bypassing event/outbox system
- No business logic in pkg/

---

## 9. Git Commit Conventions

When generating or assisting with Git commits, an `Assisted-By` trailer MUST be appended to the bottom of every commit message.
There MUST be exactly one empty line separating the body and the trailer. The format MUST strictly follow the toolchain execution path:

 - When orchestrator != executing agent: `Assisted-By: [Executing Agent] ([Model Name]) via [Orchestrator]`
   - Example: `Assisted-By: OpenCode (deep-seek-v4-pro) via Hermes`
 - When orchestrator == executing agent: `Assisted-By: [Agent] ([Model Name])`
   - Example: `Assisted-By: Hermes (deepseek-v4-pro)`

---

## 10. Design Principle

When uncertain:

> Prefer service autonomy over abstraction reuse.
