package persistence

import (
	"context"
	"fmt"

	"github.com/cubelitblade/community-v2/backend/services/post/internal/adapter/driven/persistence/ent"
	entpost "github.com/cubelitblade/community-v2/backend/services/post/internal/adapter/driven/persistence/ent/post"
	"github.com/cubelitblade/community-v2/backend/services/post/internal/domain/model"
	"github.com/cubelitblade/community-v2/backend/services/post/internal/domain/port"
)

type PostSaver struct {
	client *ent.Client
}

var _ port.PostSaver = (*PostSaver)(nil)

func NewPostSaver(client *ent.Client) port.PostSaver {
	return &PostSaver{
		client: client,
	}
}

func (w *PostSaver) Save(ctx context.Context, val model.Post) error {
	client := w.client
	if tx := ent.TxFromContext(ctx); tx != nil {
		client = tx.Client()
	}

	_, err := client.Post.Create().
		SetID(val.ID()).
		SetAuthorID(val.AuthorID()).
		SetNillableTitle(val.Title()).
		SetContent(val.Content()).
		SetStatus(entpost.Status(val.Status())).
		SetCreatedAt(val.CreatedAt()).
		SetUpdatedAt(val.UpdatedAt()).
		Save(ctx)
	if err != nil {
		return fmt.Errorf("publish post: %w", err)
	}

	return nil
}
