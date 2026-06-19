// Package config defines the configs used across the authz service.
package config

import "go.uber.org/fx"

// Config represents the top-level configuration for the authz service.
type Config struct {
	fx.Out

	OpenFGA  OpenFGAConfig  `koanf:"openfga"`
	RabbitMQ RabbitMQConfig `koanf:"rabbitmq"`
	Slog     SlogConfig     `koanf:"slog"`
	GRPC     GRPCConfig     `koanf:"grpc"`
	OTEL     OTELConfig     `koanf:"otel"`
}
