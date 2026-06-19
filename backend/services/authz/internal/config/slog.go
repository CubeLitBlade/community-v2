package config

// SlogConfig holds the configuration for the slog structured logger.
type SlogConfig struct {
	Level   string         `koanf:"level"`
	Handler string         `koanf:"handler"`
	Tint    SlogTintConfig `koanf:"tint"`
}

// SlogTintConfig holds the configuration specific to the tint slog handler.
type SlogTintConfig struct {
	TimeFormat string `koanf:"time_format"`
}
