// Package outbox implements the transactional outbox pattern for reliable event publishing.
// It provides a Relay component that periodically polls the database for pending entries,
// transforms them into CloudEvents, and publishes them to an external message broker.
package outbox

import "time"

// EntryOption applies a configuration to an Entry.
type EntryOption func(*Entry)

// Entry represents a single outbox message waiting to be published.
type Entry struct {
	ID            int64
	AggregateID   int64
	AggregateType string
	Topic         string
	EventType     string
	Payload       any
	CreatedAt     time.Time
	PublishedAt   *time.Time
	TraceID       *string
}

// NewEntry creates a new Entry with the required fields and applies any provided Options.
func NewEntry(
	id int64, aggregateType string, topic string, eventType string, now time.Time, opts ...EntryOption,
) *Entry {
	event := &Entry{
		ID:            id,
		AggregateID:   id,
		AggregateType: aggregateType,
		Topic:         topic,
		EventType:     eventType,
		Payload:       nil,
		CreatedAt:     now,
		PublishedAt:   nil,
		TraceID:       nil,
	}

	for _, opt := range opts {
		opt(event)
	}

	return event
}

// WithPayload sets the payload for the Entry.
func WithPayload(payload any) EntryOption {
	return func(entry *Entry) {
		entry.Payload = payload
	}
}

// WithPublishedAt sets the published timestamp for the Entry.
func WithPublishedAt(publishedAt *time.Time) EntryOption {
	return func(entry *Entry) {
		entry.PublishedAt = publishedAt
	}
}

// WithTraceID sets the trace ID for the Entry.
func WithTraceID(traceID *string) EntryOption {
	return func(entry *Entry) {
		entry.TraceID = traceID
	}
}
