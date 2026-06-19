package bootstrap

import (
	"context"

	"go.uber.org/fx"
	"google.golang.org/grpc"
	"google.golang.org/grpc/health"
	"google.golang.org/grpc/health/grpc_health_v1"

	"github.com/cubelitblade/community-v2/backend/services/authz/internal/application"
)

func NewHealthServer() *health.Server {
	return health.NewServer()
}

func RegisterHealthServer(grpcSrv *grpc.Server, healthSrv *health.Server) {
	grpc_health_v1.RegisterHealthServer(grpcSrv, healthSrv)
}

func StartHealthChecker(lc fx.Lifecycle, checker *application.HealthChecker) {
	lc.Append(fx.Hook{
		OnStart: func(ctx context.Context) error {
			return checker.Start(ctx)
		},
		OnStop: func(context.Context) error {
			checker.Stop()
			return nil
		},
	})
}

func HealthModule() fx.Option {
	return fx.Options(
		fx.Provide(NewHealthServer),

		fx.Invoke(RegisterHealthServer),
		fx.Invoke(StartHealthChecker),
	)
}
