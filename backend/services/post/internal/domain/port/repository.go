package port

import (
	"context"

	"github.com/cubelitblade/community-v2/backend/services/post/internal/domain/model"
)

type PostSaver interface {
	Save(ctx context.Context, val model.Post) error
}
