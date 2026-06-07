// Package constant defines shared constants for the authz service.
package constant

// Log field keys used for structured logging.
const (
	LogKeyService   = "service"
	LogKeyComponent = "component"
	LogKeyError     = "error"
	LogKeyEventID   = "event_id"
	LogKeyEventType = "event_type"
)

// Service name identifiers.
const (
	LogServiceAuthz = "authz"
)

// Component name identifiers used for structured logging.
const (
	LogComponentSyncer           = "syncer"
	LogComponentTransportHandler = "transport/handler"
)
