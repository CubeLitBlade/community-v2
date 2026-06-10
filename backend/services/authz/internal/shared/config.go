// Package shared defines the shared types and interfaces used across the authz service.
package shared

// Config represents the top-level configuration for the authz service.
type Config struct {
	ServiceName string `koanf:"service_name"`
	HTTPAddr    string `koanf:"http_addr"`

	OpenFGA  *OpenFGAConfig  `koanf:"openfga"`
	RabbitMQ *RabbitMQConfig `koanf:"rabbitmq"`
	Gin      *GinConfig      `koanf:"gin"`
	Slog     *SlogConfig     `koanf:"slog"`
}

// OpenFGAConfig holds the configuration for OpenFGA.
type OpenFGAConfig struct {
	APIURL  string `koanf:"api_url"`
	StoreID string `koanf:"store_id"`
	ModelID string `koanf:"model_id"`
}

// RabbitMQConfig holds the configuration for the RabbitMQ message broker connection.
type RabbitMQConfig struct {
	BrokerURL string `koanf:"broker_url"`
}

// SlogConfig holds the configuration for the slog structured logger.
type SlogConfig struct {
	Level string         `koanf:"level"`
	Type  string         `koanf:"type"`
	Tint  SlogTintConfig `koanf:"tint"`
}

// SlogTintConfig holds the configuration specific to the tint slog handler.
type SlogTintConfig struct {
	TimeFormat string `koanf:"time_format"`
}

// GinConfig holds the configuration for the Gin web framework.
type GinConfig struct {
	Mode string `koanf:"mode"`
}
