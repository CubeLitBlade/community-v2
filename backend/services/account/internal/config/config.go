package config

type Config struct {
	ServiceName string          `koanf:"service_name"`
	GRPC        GRPCConfig      `koanf:"grpc"`
	Database    DatabaseConfig  `koanf:"database"`
	Snowflake   SnowflakeConfig `koanf:"snowflake"`
	Slog        SlogConfig      `koanf:"slog"`
	OTEL        OTELConfig      `koanf:"otel"`
	RabbitMQ    RabbitMQConfig  `koanf:"rabbitmq"`
}
