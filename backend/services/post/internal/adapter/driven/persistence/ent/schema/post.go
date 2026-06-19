package schema

import (
	"time"

	"entgo.io/ent"
	"entgo.io/ent/dialect/entsql"
	"entgo.io/ent/schema/edge"
	"entgo.io/ent/schema/field"
)

// Post holds the schema definition for the Post entity.
type Post struct {
	ent.Schema
}

// Fields of the Post.
func (Post) Fields() []ent.Field {
	return []ent.Field{
		field.Int64("id"),
		field.Int64("author_id").Optional(),
		field.String("title").Nillable().Optional(),
		field.Text("content").NotEmpty(),
		field.Enum("status").Values("published", "archived").Default("published"),
		field.Time("created_at").Default(time.Now),
		field.Time("updated_at").Default(time.Now),
	}
}

// Edges of the Post.
func (Post) Edges() []ent.Edge {
	return []ent.Edge{
		edge.To("author", AccountProfile.Type).Field("author_id").Unique().Annotations(
			entsql.OnDelete(entsql.NoAction),
		),
	}
}
