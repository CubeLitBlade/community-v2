// Package bootstrap is responsible for assembling the application's dependency graph and managing its lifecycle.
//
// It wires together infrastructure components (gRPC, RabbitMQ, OpenTelemetry), adapters, and use cases using the
// uber/fx dependency injection framework.
package bootstrap
