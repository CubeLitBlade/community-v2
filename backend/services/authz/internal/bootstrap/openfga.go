package bootstrap

import (
	"context"
	"fmt"

	openfgav1 "github.com/openfga/api/proto/openfga/v1"
	"go.opentelemetry.io/contrib/instrumentation/google.golang.org/grpc/otelgrpc"
	"go.opentelemetry.io/otel/sdk/trace"
	"go.uber.org/fx"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
	"google.golang.org/grpc/health/grpc_health_v1"

	"github.com/cubelitblade/community-v2/backend/services/authz/internal/adapter/driven/openfga"
	"github.com/cubelitblade/community-v2/backend/services/authz/internal/config"
)

func NewOpenFGAServiceConn(
	cfg config.OpenFGAConfig, lc fx.Lifecycle, tp *trace.TracerProvider,
) (*grpc.ClientConn, error) {
	conn, err := grpc.NewClient(cfg.APIURL,
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

func NewOpenFGAServiceClient(conn *grpc.ClientConn) openfgav1.OpenFGAServiceClient {
	return openfgav1.NewOpenFGAServiceClient(conn)
}

func NewOpenFGAHealthClient(conn *grpc.ClientConn) grpc_health_v1.HealthClient {
	return grpc_health_v1.NewHealthClient(conn)
}

func OpenFGAModule() fx.Option {
	return fx.Options(
		fx.Provide(openfga.NewWriter),
		fx.Provide(openfga.NewChecker),

		fx.Provide(fx.Annotate(openfga.NewHealthChecker, fx.ResultTags(`group:"health_checkers"`))),

		fx.Provide(NewOpenFGAServiceConn),
		fx.Provide(NewOpenFGAHealthClient),
		fx.Provide(NewOpenFGAServiceClient),
	)
}
