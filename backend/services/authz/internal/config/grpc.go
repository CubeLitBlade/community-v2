package config

// GRPCConfig holds the configuration for the gRPC server.
type GRPCConfig struct {
	Addr string `koanf:"addr"`

	// Reflection enables gRPC server reflection for debugging tools.
	Reflection bool `koanf:"reflection"`
}
