package config

type GRPCConfig struct {
	Addr       string `koanf:"addr"`
	Reflection bool   `koanf:"reflection"`
}
