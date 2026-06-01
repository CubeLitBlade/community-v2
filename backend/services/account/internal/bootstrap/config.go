package bootstrap

import (
	"errors"
	"fmt"
	"log/slog"
	"net/http"
	"strings"
	"time"

	"github.com/caarlos0/env"
	platformconfig "github.com/cubelitblade/community-v2/backend/pkg/platform/config"
	authTransport "github.com/cubelitblade/community-v2/backend/services/account/internal/auth/transport"
	"github.com/gin-gonic/gin"
	"go.uber.org/fx"
	"gorm.io/gorm/logger"
)

// AppEnv represents the application deployment environment.
type AppEnv string

const (
	envDevelopment AppEnv = "dev"
	envProduction  AppEnv = "prod"
)

const defaultSlowThreshold = 200 * time.Millisecond

// Sentinel errors for config validation.
var (
	ErrAccessTTLNotPositive      = errors.New("jwt access token ttl must be positive")
	ErrJWTSecretTooShort         = errors.New("jwt secret must be at least 32 bytes")
	ErrAccessCookieNameEmpty     = errors.New("access token cookie name must not be empty")
	ErrProductionCookieNotSecure = errors.New("cookie secure must be true in production")
)

// RabbitMQConfig holds the configuration for connecting to RabbitMQ and publishing events.
type RabbitMQConfig struct {
	BrokerURL    string
	ExchangeName string
}

// JWTConfig holds the configuration for JWT signing and parsing.
type JWTConfig struct {
	Key      string
	Issuer   string
	Validity time.Duration
}

// Config holds all application configuration grouped by concern.
type Config struct {
	Slog       *platformconfig.SlogConfig
	GormLogger *platformconfig.GormLoggerConfig
	DB         *platformconfig.DBConfig
	Gin        *platformconfig.GinConfig
	Server     *platformconfig.ServerConfig
	Snowflake  *platformconfig.SnowflakeConfig
	JWT        *JWTConfig
	AuthCookie *authTransport.CookieConfig
	Publisher  *RabbitMQConfig
}

// ConfigOut provides all config sub-structs as separate fx dependencies.
type ConfigOut struct {
	fx.Out

	Slog       *platformconfig.SlogConfig
	GormLogger *platformconfig.GormLoggerConfig
	DB         *platformconfig.DBConfig
	Gin        *platformconfig.GinConfig
	Server     *platformconfig.ServerConfig
	Snowflake  *platformconfig.SnowflakeConfig
	JWT        *JWTConfig
	AuthCookie *authTransport.CookieConfig
	Publisher  *RabbitMQConfig
}

// RawConfig is the flat env-var representation parsed by caarlos0/env.
type RawConfig struct {
	AppEnv                AppEnv        `env:"APP_ENV" envDefault:"dev"`
	HTTPAddr              string        `env:"HTTP_ADDR" envDefault:":8080"`
	DSN                   string        `env:"DATABASE_URL" required:"true"`
	AccessTokenCookieName string        `env:"ACCESS_TOKEN_COOKIE_NAME" envDefault:"access_token"`
	JWTSecret             string        `env:"JWT_SECRET" required:"true"`
	SnowflakeID           int64         `env:"SNOWFLAKE_ID" required:"true"`
	AccessTokenTTL        time.Duration `env:"ACCESS_TOKEN_TTL" envDefault:"2h"`
	CookieSameSite        string        `env:"COOKIE_SAME_SITE" envDefault:"lax"`
	CookieSecure          bool          `env:"COOKIE_SECURE" envDefault:"false"`
	RabbitMQURL           string        `env:"RMQ_URL" required:"true"`
	RabbitMQExchangeName  string        `env:"RMQ_SYSTEM_EXCHANGE_NAME" required:"true"`
}

// LoadConfig reads environment variables and returns a validated Config.
func LoadConfig() (*Config, error) {
	var raw RawConfig

	err := env.Parse(&raw)
	if err != nil {
		return nil, fmt.Errorf("failed to parse config: %w", err)
	}

	jwtCfg, err := raw.newJWTConfig()
	if err != nil {
		return nil, fmt.Errorf("failed to create JWT config: %w", err)
	}

	cookieCfg, err := raw.newAuthCookieConfig()
	if err != nil {
		return nil, fmt.Errorf("failed to create auth cookie config: %w", err)
	}

	return &Config{
		Slog:       raw.newSlogConfig(),
		GormLogger: raw.newGormLoggerConfig(),
		DB:         raw.newDBConfig(),
		Gin:        raw.newGinConfig(),
		Server:     raw.newServerConfig(),
		Snowflake:  raw.newSnowflakeConfig(),
		JWT:        jwtCfg,
		AuthCookie: cookieCfg,
		Publisher:  raw.newPublisherConfig(),
	}, nil
}

// ProvideConfig loads configuration from the environment and exposes each
// config struct as a separate fx dependency via ConfigOut.
func ProvideConfig() (ConfigOut, error) {
	cfg, err := LoadConfig()
	if err != nil {
		return ConfigOut{}, fmt.Errorf("failed to load config: %w", err)
	}

	return ConfigOut{
		Out:        fx.Out{},
		Slog:       cfg.Slog,
		GormLogger: cfg.GormLogger,
		DB:         cfg.DB,
		Gin:        cfg.Gin,
		Server:     cfg.Server,
		Snowflake:  cfg.Snowflake,
		JWT:        cfg.JWT,
		AuthCookie: cfg.AuthCookie,
		Publisher:  cfg.Publisher,
	}, nil
}

func (r *RawConfig) newSlogConfig() *platformconfig.SlogConfig {
	switch r.AppEnv {
	case envDevelopment:
		return &platformconfig.SlogConfig{
			LogLevel:  slog.LevelDebug,
			LogFormat: platformconfig.LogFormatText,
		}
	case envProduction:
		return &platformconfig.SlogConfig{
			LogLevel:  slog.LevelInfo,
			LogFormat: platformconfig.LogFormatJSON,
		}
	default:
		return &platformconfig.SlogConfig{
			LogLevel:  slog.LevelDebug,
			LogFormat: platformconfig.LogFormatText,
		}
	}
}

func (r *RawConfig) newGormLoggerConfig() *platformconfig.GormLoggerConfig {
	switch r.AppEnv {
	case envDevelopment:
		return &platformconfig.GormLoggerConfig{
			GormLogLevel:  logger.Warn,
			SlowThreshold: defaultSlowThreshold,
			Colorful:      true,
		}
	case envProduction:
		return &platformconfig.GormLoggerConfig{
			GormLogLevel:  logger.Info,
			SlowThreshold: defaultSlowThreshold,
			Colorful:      true,
		}
	default:
		return &platformconfig.GormLoggerConfig{
			GormLogLevel:  logger.Warn,
			SlowThreshold: defaultSlowThreshold,
			Colorful:      true,
		}
	}
}

func (r *RawConfig) newDBConfig() *platformconfig.DBConfig {
	return &platformconfig.DBConfig{
		DSN: r.DSN,
	}
}

func (r *RawConfig) newGinConfig() *platformconfig.GinConfig {
	switch r.AppEnv {
	case envDevelopment:
		return &platformconfig.GinConfig{
			GinMode:        gin.DebugMode,
			TrustedProxies: nil,
		}
	case envProduction:
		return &platformconfig.GinConfig{
			GinMode:        gin.ReleaseMode,
			TrustedProxies: nil,
		}
	default:
		return &platformconfig.GinConfig{
			GinMode:        gin.DebugMode,
			TrustedProxies: nil,
		}
	}
}

func (r *RawConfig) newServerConfig() *platformconfig.ServerConfig {
	return &platformconfig.ServerConfig{
		HTTPAddr:          r.HTTPAddr,
		ReadHeaderTimeout: 5 * time.Second,
		ReadTimeout:       5 * time.Second,
		WriteTimeout:      5 * time.Second,
	}
}

func (r *RawConfig) newSnowflakeConfig() *platformconfig.SnowflakeConfig {
	return &platformconfig.SnowflakeConfig{
		NodeID: r.SnowflakeID,
	}
}

func (r *RawConfig) newJWTConfig() (*JWTConfig, error) {
	if r.AccessTokenTTL <= 0 {
		return nil, ErrAccessTTLNotPositive
	}

	//nolint:mnd // 32 bytes is the minimum key length for HMAC-SHA256 (HS256)
	if len(r.JWTSecret) < 32 {
		return nil, ErrJWTSecretTooShort
	}

	return &JWTConfig{
		Key:      r.JWTSecret,
		Issuer:   "community-v2",
		Validity: r.AccessTokenTTL,
	}, nil
}

func (r *RawConfig) newAuthCookieConfig() (*authTransport.CookieConfig, error) {
	if r.AccessTokenCookieName == "" {
		return nil, ErrAccessCookieNameEmpty
	}

	if r.AppEnv == envProduction && !r.CookieSecure {
		return nil, ErrProductionCookieNotSecure
	}

	var sameSite http.SameSite

	switch strings.ToLower(r.CookieSameSite) {
	case "lax":
		sameSite = http.SameSiteLaxMode
	case "none":
		sameSite = http.SameSiteNoneMode
	case "strict":
		sameSite = http.SameSiteStrictMode
	}

	return &authTransport.CookieConfig{
		Name:     r.AccessTokenCookieName,
		Secure:   r.CookieSecure,
		SameSite: sameSite,
		MaxAge:   int(r.AccessTokenTTL.Seconds()),
	}, nil
}

func (r *RawConfig) newPublisherConfig() *RabbitMQConfig {
	return &RabbitMQConfig{
		BrokerURL:    r.RabbitMQURL,
		ExchangeName: r.RabbitMQExchangeName,
	}
}
