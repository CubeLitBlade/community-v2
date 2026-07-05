package bootstrap

import (
	"context"
	"fmt"
	"log/slog"
	"net"
	"strings"

	"go.opentelemetry.io/contrib/instrumentation/google.golang.org/grpc/otelgrpc"
	"go.opentelemetry.io/otel/metric"
	"go.opentelemetry.io/otel/trace"
	"go.uber.org/fx"
	"google.golang.org/grpc"
	"google.golang.org/grpc/reflection"
	"google.golang.org/grpc/stats"

	authzv1 "github.com/cubelitblade/community-v2/backend/gen/go/authz/v1"
	authgrpc "github.com/cubelitblade/community-v2/backend/services/authz/internal/adapter/driving/grpc"
	"github.com/cubelitblade/community-v2/backend/services/authz/internal/config"
)

type GRPCServerParam struct {
	fx.In

	CFG      config.GRPCConfig
	LC       fx.Lifecycle
	AuthzSrv *authgrpc.AuthzServiceServer
	TP       trace.TracerProvider
	MP       metric.MeterProvider
	Logger   *slog.Logger
}

func NewGRPCServer(params GRPCServerParam) (*grpc.Server, error) {
	srv := grpc.NewServer(
		grpc.StatsHandler(otelgrpc.NewServerHandler(
			otelgrpc.WithMeterProvider(params.MP),
			otelgrpc.WithTracerProvider(params.TP),
			otelgrpc.WithFilter(func(ri *stats.RPCTagInfo) bool {
				return !strings.HasSuffix(ri.FullMethodName, "/grpc.health.v1.Health/Check")
			}),
		)),
	)

	authzv1.RegisterAuthzServiceServer(srv, params.AuthzSrv)

	if params.CFG.Reflection {
		reflection.Register(srv)
	}

	params.LC.Append(fx.Hook{
		OnStart: func(ctx context.Context) error {
			listenCfg := net.ListenConfig{}

			lis, err := listenCfg.Listen(ctx, "tcp", params.CFG.Addr)
			if err != nil {
				return fmt.Errorf("failed to listen: %w", err)
			}

			go func() {
				params.Logger.InfoContext(ctx, "gRPC server started", slog.String("addr", lis.Addr().String()))

				if err := srv.Serve(lis); err != nil {
					params.Logger.ErrorContext(ctx, "gRPC server failed", slog.Any("error", err))
				}
			}()

			return nil
		},
		OnStop: func(ctx context.Context) error {
			params.Logger.InfoContext(ctx, "stopping gRPC server gracefully")
			srv.GracefulStop()

			return nil
		},
	})

	return srv, nil
}

func GRPCModule() fx.Option {
	return fx.Options(
		fx.Provide(NewGRPCServer),
		fx.Provide(authgrpc.NewAuthzServiceServer),
		fx.Provide(authgrpc.NewGRPCHealthReporter),
	)
}
