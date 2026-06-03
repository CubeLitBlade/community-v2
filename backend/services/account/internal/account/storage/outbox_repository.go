package storage

import (
	"context"
	"fmt"
	"time"

	"github.com/cubelitblade/community-v2/backend/pkg/events/outbox"
	"gorm.io/cli/gorm/field"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

// OutboxRepository implements the outbox.Scanner, outbox.Recorder, and account.OutboxSaver interfaces.
type OutboxRepository struct{}

// EventRow represents a record in the "outbox" PostgreSQL table.
type EventRow struct {
	ID            int64      `gorm:"column:id;primary_key"`
	AggregateID   int64      `gorm:"column:aggregate_id"`
	AggregateType string     `gorm:"column:aggregate_type"`
	EventType     string     `gorm:"column:event_type"`
	Payload       any        `gorm:"payload;type:jsonb;serializer:json"`
	CreatedAt     time.Time  `gorm:"column:created_at"`
	PublishedAt   *time.Time `gorm:"column:published_at"`
	TraceID       *string    `gorm:"column:trace_id"`
}

// EventRowField provides strongly-typed Gorm field references for the outbox table columns.
var EventRowField = struct {
	ID            field.Number[int64]
	AggregateID   field.Number[int64]
	AggregateType field.String
	EventType     field.String
	Payload       field.Bytes
	CreatedAt     field.Time
	PublishedAt   field.Time
	TraceID       field.String
}{
	ID:            field.Number[int64]{}.WithColumn("id"),
	AggregateID:   field.Number[int64]{}.WithColumn("aggregate_id"),
	AggregateType: field.String{}.WithColumn("aggregate_type"),
	EventType:     field.String{}.WithColumn("event_type"),
	Payload:       field.Bytes{}.WithColumn("payload"),
	CreatedAt:     field.Time{}.WithColumn("created_at"),
	PublishedAt:   field.Time{}.WithColumn("published_at"),
	TraceID:       field.String{}.WithColumn("trace_id"),
}

// NewOutboxRepository creates a new OutboxRepository.
func NewOutboxRepository() *OutboxRepository { return &OutboxRepository{} }

// Save persists a new outbox event to the database.
func (r *OutboxRepository) Save(ctx context.Context, db *gorm.DB, event *outbox.Entry) error {
	row := &EventRow{
		ID:            event.ID,
		AggregateID:   event.AggregateID,
		AggregateType: event.AggregateType,
		EventType:     event.EventType,
		Payload:       event.Payload,
		CreatedAt:     event.CreatedAt,
		PublishedAt:   event.PublishedAt,
		TraceID:       event.TraceID,
	}

	err := gorm.G[EventRow](db).Create(ctx, row)
	if err != nil {
		return fmt.Errorf("insert event: %w", err)
	}

	return nil
}

// Scan retrieves a limited number of unpublished outbox events, locking them for update.
func (r *OutboxRepository) Scan(ctx context.Context, db *gorm.DB, count int) ([]*outbox.Entry, error) {
	tx := db.Clauses(clause.Locking{Strength: "UPDATE", Options: "SKIP LOCKED"})

	rows, err := gorm.G[EventRow](tx).
		Where(EventRowField.PublishedAt.IsNull()).
		Order(EventRowField.CreatedAt).
		Limit(count).
		Find(ctx)
	if err != nil {
		return nil, fmt.Errorf("scan: %w", err)
	}

	var events []*outbox.Entry

	for _, row := range rows {
		event := &outbox.Entry{
			ID:            row.ID,
			AggregateID:   row.AggregateID,
			AggregateType: row.AggregateType,
			EventType:     row.EventType,
			Payload:       row.Payload,
			CreatedAt:     row.CreatedAt,
			PublishedAt:   row.PublishedAt,
			TraceID:       row.TraceID,
		}
		events = append(events, event)
	}

	return events, nil
}

// Ack marks an outbox event as published by setting its PublishedAt timestamp.
func (r *OutboxRepository) Ack(ctx context.Context, db *gorm.DB, id int64, now time.Time) error {
	_, err := gorm.G[EventRow](db).
		Where(EventRowField.ID.Eq(id)).
		Set(EventRowField.PublishedAt.Set(now)).
		Update(ctx)
	if err != nil {
		return fmt.Errorf("mark as published: %w", err)
	}

	return nil
}

// TableName overrides the default Gorm table name to "outbox".
func (EventRow) TableName() string {
	return "outbox"
}
