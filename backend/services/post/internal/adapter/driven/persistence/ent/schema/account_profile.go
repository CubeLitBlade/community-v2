package schema

import (
	"entgo.io/ent"
	"entgo.io/ent/schema/edge"
	"entgo.io/ent/schema/field"
)

// AccountProfile holds the schema definition for the AccountProfile entity.
type AccountProfile struct {
	ent.Schema
}

// Fields of the AccountProfile.
func (AccountProfile) Fields() []ent.Field {
	return []ent.Field{
		field.Int64("id"),
		field.String("username"),
		field.String("displayName"),
	}
}

// Edges of the AccountProfile.
func (AccountProfile) Edges() []ent.Edge {
	return []ent.Edge{
		edge.From("posts", Post.Type).Ref("author"),
	}
}
