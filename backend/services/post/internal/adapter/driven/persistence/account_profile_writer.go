package persistence

import (
	"context"
	"fmt"

	"github.com/cubelitblade/community-v2/backend/services/post/internal/adapter/driven/persistence/ent"
	"github.com/cubelitblade/community-v2/backend/services/post/internal/domain/model"
)

type AccountProfileWriter struct{}

func (*AccountProfileWriter) Save(ctx context.Context, client *ent.Client, val model.AccountProfile) error {
	_, err := client.AccountProfile.Create().
		SetID(val.ID).
		SetUsername(val.Username).
		SetDisplayName(val.DisplayName).
		Save(ctx)
	if err != nil {
		return fmt.Errorf("write account profile: %w", err)
	}

	return nil
}
