package persistence

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/cubelitblade/community-v2/backend/services/post/internal/adapter/driven/persistence/ent"
	"github.com/cubelitblade/community-v2/backend/services/post/internal/domain/port"
)

// OutboxRecorder persists outbox entries using ent.
type OutboxRecorder struct {
	client *ent.Client
	now    func() time.Time
}

var _ port.EventRecorder = (*OutboxRecorder)(nil)

func NewOutboxRecorder(client *ent.Client) *OutboxRecorder {
	return &OutboxRecorder{
		client: client,
		now:    time.Now,
	}
}

func (r *OutboxRecorder) Record(ctx context.Context, entry *port.OutboxEntry) error {
	client := r.client
	if tx := ent.TxFromContext(ctx); tx != nil {
		client = tx.Client()
	}

	payload, err := marshalPayload(entry.Payload)
	if err != nil {
		return fmt.Errorf("marshal outbox payload: %w", err)
	}

	_, err = client.Outbox.Create().
		SetID(entry.ID).
		SetAggregateID(entry.AggregateID).
		SetAggregateType(entry.AggregateType).
		SetTopic(entry.Topic).
		SetEventType(entry.EventType).
		SetPayload(payload).
		SetCreatedAt(r.now()).
		Save(ctx)
	if err != nil {
		return fmt.Errorf("record outbox: %w", err)
	}

	return nil
}

// marshalPayload converts a Go struct to a JSON-serializable map.
func marshalPayload(v any) (map[string]any, error) {
	switch m := v.(type) {
	case map[string]any:
		return m, nil
	default:
		bytes, err := json.Marshal(v)
		if err != nil {
			return nil, err
		}
		var out map[string]any
		if err := json.Unmarshal(bytes, &out); err != nil {
			return nil, err
		}
		return out, nil
	}
}
