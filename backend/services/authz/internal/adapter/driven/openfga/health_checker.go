package openfga

import (
	"context"
	"errors"
	"fmt"

	"google.golang.org/grpc"
	"google.golang.org/grpc/health/grpc_health_v1"

	"github.com/cubelitblade/community-v2/backend/services/authz/internal/domain/port"
)

var ErrDependencyNotServing = errors.New("dependency is not serving")

type HealthChecker struct {
	conn *grpc.ClientConn
}

var _ port.HealthChecker = (*HealthChecker)(nil)

func NewHealthChecker(conn *grpc.ClientConn) port.HealthChecker {
	return &HealthChecker{
		conn: conn,
	}
}

func (*HealthChecker) Name() string {
	return "openfga"
}

func (c *HealthChecker) Check(ctx context.Context) error {
	client := grpc_health_v1.NewHealthClient(c.conn)

	resp, err := client.Check(ctx, &grpc_health_v1.HealthCheckRequest{})
	if err != nil {
		return fmt.Errorf("openfga health rpc failed: %w", err)
	}

	if resp.GetStatus() != grpc_health_v1.HealthCheckResponse_SERVING {
		return fmt.Errorf("%w: OpenFGA status %s", ErrDependencyNotServing, resp.GetStatus())
	}

	return nil
}
