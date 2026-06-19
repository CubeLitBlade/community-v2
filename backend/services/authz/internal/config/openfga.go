package config

// OpenFGAConfig holds the configuration for OpenFGA.
type OpenFGAConfig struct {
	APIURL  string `koanf:"api_url"`
	StoreID string `koanf:"store_id"`
	ModelID string `koanf:"model_id"`
}
