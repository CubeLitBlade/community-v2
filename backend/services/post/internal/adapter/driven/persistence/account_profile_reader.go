package persistence

import (
	"context"
	"fmt"

	"github.com/cubelitblade/community-v2/backend/services/post/internal/adapter/driven/persistence/ent"
	"github.com/cubelitblade/community-v2/backend/services/post/internal/domain/model"
)

type AccountProfileReader struct{}

func (*AccountProfileReader) Get(ctx context.Context, client *ent.Client, id int64) (*model.AccountProfile, error) {
	row, err := client.AccountProfile.Get(ctx, id)
	if err != nil {
		return nil, fmt.Errorf("get account profile: %w", err)
	}

	return &model.AccountProfile{
		ID:          row.ID,
		Username:    row.Username,
		DisplayName: row.DisplayName,
	}, nil
}
