package config

type SlogConfig struct {
	Level   string         `koanf:"level"`
	Handler string         `koanf:"handler"`
	Tint    SlogTintConfig `koanf:"tint"`
}

type SlogTintConfig struct {
	TimeFormat string `koanf:"time_format"`
}
