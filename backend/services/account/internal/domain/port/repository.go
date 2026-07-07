package port

import (
	"context"

	"github.com/cubelitblade/community-v2/backend/services/account/internal/domain"
)

type AccountReader interface {
	FindByID(ctx context.Context, id int64) (*account.Account, error)
	FindByUsername(ctx context.Context, username account.Username) (*account.Account, error)
}

type AccountWriter interface {
	Create(ctx context.Context, acc *account.Account) error
	Update(ctx context.Context, acc *account.Account) error
}
