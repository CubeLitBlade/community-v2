package config

type DatabaseConfig struct {
	URI string `koanf:"uri" validate:"required,uri"`
}
