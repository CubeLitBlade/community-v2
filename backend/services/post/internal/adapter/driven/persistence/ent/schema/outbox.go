package schema

import (
	"time"

	"entgo.io/ent"
	"entgo.io/ent/dialect/entsql"
	"entgo.io/ent/schema"
	"entgo.io/ent/schema/field"
)

// Outbox holds the schema definition for outbox entries.
type Outbox struct {
	ent.Schema
}

// Annotations of the Outbox.
func (Outbox) Annotations() []schema.Annotation {
	return []schema.Annotation{
		&entsql.Annotation{
			Table: "outbox",
		},
	}
}

// Fields of the Outbox.
func (Outbox) Fields() []ent.Field {
	return []ent.Field{
		field.Int64("id"),
		field.Int64("aggregate_id"),
		field.String("aggregate_type"),
		field.String("topic"),
		field.String("event_type"),
		field.JSON("payload", map[string]any{}),
		field.Time("created_at").Default(time.Now),
		field.Time("published_at").Optional().Nillable(),
		field.String("trace_id").Optional().Nillable(),
	}
}
