package port

import "context"

type Authorizer interface {
	CanPublishPost(ctx context.Context, accountID int64) (bool, error)
}
