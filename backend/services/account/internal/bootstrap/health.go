package bootstrap

import (
	"go.uber.org/fx"
	"google.golang.org/grpc"
	"google.golang.org/grpc/health/grpc_health_v1"

	accountgrpc "github.com/cubelitblade/community-v2/backend/services/account/internal/adapter/driving/grpc"
)

func RegisterHealthServer(grpcSrv *grpc.Server, healthSrv *accountgrpc.HealthServer) {
	grpc_health_v1.RegisterHealthServer(grpcSrv, healthSrv.Server())
	healthSrv.SetServing()
}

func HealthModule() fx.Option {
	return fx.Options()
}
