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
	"google.golang.org/grpc/credentials/insecure"
	"google.golang.org/grpc/reflection"

	postv1 "github.com/cubelitblade/community-v2/backend/gen/go/post/v1"
	postgrpc "github.com/cubelitblade/community-v2/backend/services/post/internal/adapter/driving/grpc"
	"github.com/cubelitblade/community-v2/backend/services/post/internal/config"
)

func NewGRPCClientConn(lc fx.Lifecycle, tp *trace.TracerProvider) (*grpc.ClientConn, error) {
	conn, err := grpc.NewClient("localhost:50051",
		grpc.WithTransportCredentials(insecure.NewCredentials()),
		grpc.WithStatsHandler(otelgrpc.NewClientHandler(
			otelgrpc.WithTracerProvider(tp),
		)),
	)
	if err != nil {
		return nil, fmt.Errorf("creating client: %w", err)
	}

	lc.Append(fx.Hook{
		OnStop: func(context.Context) error {
			return conn.Close()
		},
	})

	return conn, nil
}

func NewGRPCServer(
	cfg config.GRPCConfig, lc fx.Lifecycle, postSrv *postgrpc.PostServiceServer, tp *trace.TracerProvider,
	logger *slog.Logger,
) (*grpc.Server, error) {
	srv := grpc.NewServer(
		grpc.StatsHandler(otelgrpc.NewServerHandler(
			otelgrpc.WithTracerProvider(tp),
		)),
	)

	postv1.RegisterPostServiceServer(srv, postSrv)

	if cfg.Reflection {
		reflection.Register(srv)
	}

	lc.Append(fx.Hook{
		OnStart: func(ctx context.Context) error {
			lc := net.ListenConfig{}

			lis, err := lc.Listen(ctx, "tcp", cfg.Addr)
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
		fx.Provide(postgrpc.NewPostServiceServer),

		fx.Provide(NewGRPCClientConn),

		fx.Invoke(func(*grpc.Server) {}),
	)
}
