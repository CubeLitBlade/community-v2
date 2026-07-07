package persistence

import (
	"context"
	"fmt"
	"time"

	"gorm.io/gorm"

	"github.com/cubelitblade/community-v2/backend/pkg/events/outbox"
)

type OutboxRepository struct {
	db *gorm.DB
}

func NewOutboxRepository(db *gorm.DB) *OutboxRepository {
	return &OutboxRepository{db: db}
}

type OutboxRow struct {
	ID            int64      `gorm:"column:id;primaryKey"`
	AggregateID   int64      `gorm:"column:aggregate_id"`
	AggregateType string     `gorm:"column:aggregate_type"`
	Topic         string     `gorm:"column:topic"`
	EventType     string     `gorm:"column:event_type"`
	Payload       any        `gorm:"column:payload;type:jsonb;serializer:json"`
	CreatedAt     time.Time  `gorm:"column:created_at"`
	PublishedAt   *time.Time `gorm:"column:published_at"`
	TraceID       *string    `gorm:"column:trace_id"`
}

func (OutboxRow) TableName() string {
	return "outbox"
}

func (r *OutboxRepository) Record(ctx context.Context, entry *outbox.Entry) error {
	row := &OutboxRow{
		ID:            entry.ID,
		AggregateID:   entry.AggregateID,
		AggregateType: entry.AggregateType,
		Topic:         entry.Topic,
		EventType:     entry.EventType,
		Payload:       entry.Payload,
		CreatedAt:     entry.CreatedAt,
		PublishedAt:   entry.PublishedAt,
		TraceID:       entry.TraceID,
	}

	err := txFromContext(ctx, r.db).Create(row).Error
	if err != nil {
		return fmt.Errorf("insert outbox event: %w", err)
	}

	return nil
}
