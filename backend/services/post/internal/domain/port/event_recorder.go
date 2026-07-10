package port

import "context"

// EventRecorder persists outbox entries within the same transaction as the business operation.
type EventRecorder interface {
	Record(ctx context.Context, entry *OutboxEntry) error
}

// OutboxEntry represents a single outbox message waiting to be published.
type OutboxEntry struct {
	ID            int64
	AggregateID   int64
	AggregateType string
	Topic         string
	EventType     string
	Payload       any
}
