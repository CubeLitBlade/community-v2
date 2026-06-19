package port

import "context"

type HealthChecker interface {
	Name() string
	Check(ctx context.Context) error
}
