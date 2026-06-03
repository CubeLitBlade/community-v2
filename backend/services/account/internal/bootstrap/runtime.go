package bootstrap

// AppEnv represents the application deployment environment.
type AppEnv string

// Application environment constants.
const (
	// AppEnvDevelopment is the development environment.
	AppEnvDevelopment AppEnv = "dev"
	// AppEnvProduction is the production environment.
	AppEnvProduction AppEnv = "prod"
)

// RuntimeConfig holds the runtime configuration loaded from environment variables.
type RuntimeConfig struct {
	AppEnv AppEnv `env:"APP_ENV" envDefault:"dev"`
	Addr   string `env:"HTTP_ADDR" envDefault:":8080"`
}

// Environment returns the current application environment.
func (cfg RuntimeConfig) Environment() AppEnv {
	return cfg.AppEnv
}

// HTTPAddr returns the HTTP listen address.
func (cfg RuntimeConfig) HTTPAddr() string {
	return cfg.Addr
}

// AppRuntime provides read-only access to the runtime configuration.
type AppRuntime interface {
	Environment() AppEnv
	HTTPAddr() string
}
