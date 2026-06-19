package grpc

import (
	"google.golang.org/grpc/health"
	"google.golang.org/grpc/health/grpc_health_v1"

	"github.com/cubelitblade/community-v2/backend/services/authz/internal/application"
)

type HealthMonitor struct {
	server *health.Server
}

func NewGRPCHealthReporter(server *health.Server) application.ServingStatusSetter {
	return &HealthMonitor{
		server: server,
	}
}

func (r *HealthMonitor) SetStatus(status application.ServingStatus) {
	switch status {
	case application.StatusServing:
		r.server.SetServingStatus("", grpc_health_v1.HealthCheckResponse_SERVING)
	case application.StatusNotServing:
		r.server.SetServingStatus("", grpc_health_v1.HealthCheckResponse_NOT_SERVING)
	default:
		r.server.SetServingStatus("", grpc_health_v1.HealthCheckResponse_UNKNOWN)
	}
}
