# Backend AI Agent Rules

This repository is a Go monorepo backend system located under `backend/`.

It consists of multiple microservices and shared infrastructure packages.

---

## 1. Architecture Overview

The system follows a **hexagonal architecture direction**.

- Domain logic: `internal/domain`
- Adapters (DB / external systems): `internal/adapter`
- Transport layer: `internal/transport`
- Bootstrap: dependency wiring only

Legacy code may still exist (notably in Account service), but new code MUST follow hexagonal structure.

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

- Each service owns its own API and event definitions
- Contracts are NOT centralized
- Cross-service communication relies on versioned schemas

Semantic compatibility is required, structural differences are allowed.

---

## 5. Shared Packages (pkg/)

`pkg/` contains infrastructure-only code.

### Allowed:
- platform (server, DB, routing)
- events (outbox + RabbitMQ)
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
