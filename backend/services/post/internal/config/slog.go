package config

type SlogConfig struct {
	Handler string         `koanf:"handler" validate:"oneof=text json tint default discard"`
	Level   string         `koanf:"level" validate:"oneof=debug info warn error"`
	Tint    SlogTintConfig `koanf:"tint"`
}

type SlogTintConfig struct {
	TimeFormat string `koanf:"time_format"`
}
