package platform

import "google.golang.org/grpc"

type GRPCRegistrar interface {
	RegisterGRPC(s *grpc.Server)
}
