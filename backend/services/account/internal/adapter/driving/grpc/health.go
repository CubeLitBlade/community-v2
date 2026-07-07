package grpc

import (
	"google.golang.org/grpc/health"
	"google.golang.org/grpc/health/grpc_health_v1"
)

type HealthServer struct {
	server *health.Server
}

func NewHealthServer() *HealthServer {
	return &HealthServer{
		server: health.NewServer(),
	}
}

func (h *HealthServer) Server() *health.Server {
	return h.server
}

func (h *HealthServer) SetServing() {
	h.server.SetServingStatus("", grpc_health_v1.HealthCheckResponse_SERVING)
}

func (h *HealthServer) SetNotServing() {
	h.server.SetServingStatus("", grpc_health_v1.HealthCheckResponse_NOT_SERVING)
}
