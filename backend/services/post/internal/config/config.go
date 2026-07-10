// Package config defines the configs used across the post service.
package config

import (
	"go.uber.org/fx"
)

type AppEnv string

type Config struct {
	fx.Out

	Env       AppEnv          `koanf:"env" validate:"required,oneof=dev prod"`
	Database  DatabaseConfig  `koanf:"db"`
	Slog      SlogConfig      `koanf:"slog"`
	GRPC      GRPCConfig      `koanf:"grpc"`
	Snowflake SnowflakeConfig `koanf:"snowflake"`
	OTEL      OTELConfig      `koanf:"otel"`
	RabbitMQ  RabbitMQConfig  `koanf:"rabbitmq"`
}
