package bootstrap

import (
	"context"
	"fmt"
	"log/slog"
	"net"

	"go.opentelemetry.io/contrib/instrumentation/google.golang.org/grpc/otelgrpc"
	"go.opentelemetry.io/otel/sdk/trace"
	"go.uber.org/fx"
	"google.golang.org/grpc"
	"google.golang.org/grpc/reflection"

	authzv1 "github.com/cubelitblade/community-v2/backend/gen/go/authz/v1"
	authgrpc "github.com/cubelitblade/community-v2/backend/services/authz/internal/adapter/driving/grpc"
	"github.com/cubelitblade/community-v2/backend/services/authz/internal/config"
)

func NewGRPCServer(
	cfg config.GRPCConfig, lc fx.Lifecycle, authzSrv *authgrpc.AuthzServiceServer, tp *trace.TracerProvider,
	logger *slog.Logger,
) (*grpc.Server, error) {
	srv := grpc.NewServer(
		grpc.StatsHandler(otelgrpc.NewServerHandler(
			otelgrpc.WithTracerProvider(tp),
		)),
	)

	authzv1.RegisterAuthzServiceServer(srv, authzSrv)

	if cfg.Reflection {
		reflection.Register(srv)
	}

	lc.Append(fx.Hook{
		OnStart: func(ctx context.Context) error {
			listenCfg := net.ListenConfig{}

			lis, err := listenCfg.Listen(ctx, "tcp", cfg.Addr)
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
		},
		OnStop: func(ctx context.Context) error {
			logger.InfoContext(ctx, "stopping gRPC server gracefully")
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
