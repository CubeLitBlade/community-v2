package persistence

import (
	"context"
	"fmt"

	"github.com/cubelitblade/community-v2/backend/services/post/internal/adapter/driven/persistence/ent"
	"github.com/cubelitblade/community-v2/backend/services/post/internal/domain/model"
)

type PostReader struct{}

func (*PostReader) Get(ctx context.Context, client *ent.Client, id int64) (*model.Post, error) {
	row, err := client.Post.Get(ctx, id)
	if err != nil {
		return nil, fmt.Errorf("get post: %w", err)
	}

	return model.Unmarshal(
		row.ID, row.AuthorID, row.Title, row.Content, model.Status(row.Status), row.CreatedAt, row.UpdatedAt,
	), nil
}
