package port

import (
	"context"

	"github.com/cubelitblade/community-v2/backend/services/authz/internal/domain"
)

type Checker interface {
	Check(ctx context.Context, user, relation, object string) (bool, error)
}

type TupleWriter interface {
	Write(ctx context.Context, tuples []domain.Tuple) error
}
