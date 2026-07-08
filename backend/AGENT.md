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

## 6. Persistence Strategy (Learning Context)

This repository intentionally uses different persistence approaches:

- Some services use GORM
- Some services use ent

Each service MUST use only one persistence approach internally.

---

## 7. Forbidden Patterns

- No new service-to-service HTTP calls
- No shared domain models across services
- No bypassing event/outbox system
- No business logic in pkg/

---

## 8. Design Principle

When uncertain:

> Prefer service autonomy over abstraction reuse.
