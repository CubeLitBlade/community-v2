// Package telemetry provides observability decorators for the authz service.
//
// Each decorator wraps a port interface or an application service, adding
// OpenTelemetry tracing and metrics instrumentation without polluting the
// underlying business logic. Decorators implement the same interface as their
// inner component, making them transparent to all consumers.
//
// This package intentionally has no business logic and no adapter protocol
// concerns—it exists solely as a cross-cutting observability layer that is
// wired together in the bootstrap package.
package telemetry
