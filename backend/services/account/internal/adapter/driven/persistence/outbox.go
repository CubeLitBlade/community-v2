package persistence

import (
	"context"
	"fmt"
	"time"

	"gorm.io/gorm"
	"gorm.io/gorm/clause"

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

func (r *OutboxRepository) Scan(ctx context.Context, db *gorm.DB, count int) ([]*outbox.Entry, error) {
	tx := db.WithContext(ctx).Clauses(clause.Locking{Strength: "UPDATE", Options: "SKIP LOCKED"})

	rows, err := gorm.G[OutboxRow](tx).
		Where("published_at IS NULL").
		Order("created_at").
		Limit(count).
		Find(ctx)
	if err != nil {
		return nil, fmt.Errorf("scan outbox: %w", err)
	}

	entries := make([]*outbox.Entry, 0, len(rows))
	for _, row := range rows {
		entries = append(entries, &outbox.Entry{
			ID:            row.ID,
			AggregateID:   row.AggregateID,
			AggregateType: row.AggregateType,
			Topic:         row.Topic,
			EventType:     row.EventType,
			Payload:       row.Payload,
			CreatedAt:     row.CreatedAt,
			PublishedAt:   row.PublishedAt,
			TraceID:       row.TraceID,
		})
	}

	return entries, nil
}

func (r *OutboxRepository) Ack(ctx context.Context, db *gorm.DB, id int64, now time.Time) error {
	_, err := gorm.G[OutboxRow](db.WithContext(ctx)).
		Where("id = ?", id).
		Update(ctx, "published_at", now)
	if err != nil {
		return fmt.Errorf("ack outbox event: %w", err)
	}

	return nil
}

var (
	_ outbox.Scanner  = (*OutboxRepository)(nil)
	_ outbox.Recorder = (*OutboxRepository)(nil)
)
