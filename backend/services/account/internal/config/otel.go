package config

type OTELConfig struct {
	ServiceName string `koanf:"service_name"`
	Endpoint    string `koanf:"endpoint"`
	Insecure    bool   `koanf:"insecure"`
	Enable      bool   `koanf:"enable"`
}
