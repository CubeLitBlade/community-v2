package port

import "context"

type Transactor interface {
	InTx(ctx context.Context, fn func(context.Context) error) error
}
