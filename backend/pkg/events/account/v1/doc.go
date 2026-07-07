// Package v1 defines the canonical event contracts for the account service.
//
// These types and constants are the shared contract between the account
// service (publisher) and any service that consumes account events (e.g. authz).
//
// All events follow the CloudEvents specification and use reverse-DNS event
// type naming with semantic versioning.
package v1
