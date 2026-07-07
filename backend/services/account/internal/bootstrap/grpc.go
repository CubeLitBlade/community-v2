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

	accountgrpc "github.com/cubelitblade/community-v2/backend/services/account/internal/adapter/driving/grpc"
	"github.com/cubelitblade/community-v2/backend/services/account/internal/config"
)

type GRPCServerParam struct {
	fx.In

	CFG      config.GRPCConfig
	LC       fx.Lifecycle
	AccSrv   *accountgrpc.AccountServiceServer
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

	params.AccSrv.RegisterGRPC(srv)

	if params.CFG.Reflection {
		reflection.Register(srv)
	}

	params.LC.Append(fx.Hook{
		OnStart: func(ctx context.Context) error {
			return startGRPC(ctx, srv, params.CFG.Addr, params.Logger)
		},
		OnStop: func(ctx context.Context) error {
			params.Logger.InfoContext(ctx, "stopping gRPC server gracefully")
			srv.GracefulStop()

			return nil
		},
	})

	return srv, nil
}

func startGRPC(ctx context.Context, srv *grpc.Server, addr string, logger *slog.Logger) error {
	lis, err := (&net.ListenConfig{}).Listen(ctx, "tcp", addr)
	if err != nil {
		return fmt.Errorf("failed to listen: %w", err)
	}

	go func() {
		logger.InfoContext(ctx, "gRPC server started", slog.String("addr", lis.Addr().String()))

		if err := srv.Serve(lis); err != nil {
			logger.ErrorContext(ctx, "gRPC server failed", slog.Any("error", err))
		}
	}()

	return nil
}

func GRPCModule() fx.Option {
	return fx.Options(
		fx.Provide(NewGRPCServer),
		fx.Provide(accountgrpc.NewAccountServiceServer),
		fx.Provide(accountgrpc.NewHealthServer),

		fx.Invoke(RegisterHealthServer),
	)
}
