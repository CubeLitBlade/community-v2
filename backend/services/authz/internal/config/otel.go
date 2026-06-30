package config

// OTELConfig holds the configuration for OpenTelemetry tracing.
type OTELConfig struct {
	ServiceName string `koanf:"service_name"`

	// Endpoint is the OTLP exporter target address (e.g., "localhost:4317").
	Endpoint string `koanf:"endpoint"`

	// Insecure disables TLS for the telemetry connection.
	Insecure bool `koanf:"insecure"`

	// Enable toggles OpenTelemetry tracing globally.
	Enable bool `koanf:"enable"`
}
