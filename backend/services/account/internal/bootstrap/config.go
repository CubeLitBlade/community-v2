package bootstrap

import (
	"errors"
	"fmt"
	"log/slog"
	"net/http"
	"strings"
	"time"

	"github.com/caarlos0/env"
	"github.com/cubelitblade/community-v2/backend/pkg/common/jwt"
	"github.com/cubelitblade/community-v2/backend/pkg/platform"
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

// Config holds all application configuration grouped by concern.
type Config struct {
	Slog       *platform.SlogConfig
	GormLogger *platform.GormLoggerConfig
	DB         *platform.DBConfig
	Gin        *platform.GinConfig
	Server     *platform.ServerConfig
	Snowflake  *platform.SnowflakeConfig
	JWT        *jwt.Config
	AuthCookie *authTransport.CookieConfig
}

// ConfigOut provides all config sub-structs as separate fx dependencies.
type ConfigOut struct {
	fx.Out

	Slog       *platform.SlogConfig
	GormLogger *platform.GormLoggerConfig
	DB         *platform.DBConfig
	Gin        *platform.GinConfig
	Server     *platform.ServerConfig
	Snowflake  *platform.SnowflakeConfig
	JWT        *jwt.Config
	AuthCookie *authTransport.CookieConfig
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
	}, nil
}

func (r *RawConfig) newSlogConfig() *platform.SlogConfig {
	switch r.AppEnv {
	case envDevelopment:
		return &platform.SlogConfig{
			LogLevel:  slog.LevelDebug,
			LogFormat: platform.LogFormatText,
		}
	case envProduction:
		return &platform.SlogConfig{
			LogLevel:  slog.LevelInfo,
			LogFormat: platform.LogFormatJSON,
		}
	default:
		return &platform.SlogConfig{
			LogLevel:  slog.LevelDebug,
			LogFormat: platform.LogFormatText,
		}
	}
}

func (r *RawConfig) newGormLoggerConfig() *platform.GormLoggerConfig {
	switch r.AppEnv {
	case envDevelopment:
		return &platform.GormLoggerConfig{
			GormLogLevel:  logger.Warn,
			SlowThreshold: defaultSlowThreshold,
			Colorful:      true,
		}
	case envProduction:
		return &platform.GormLoggerConfig{
			GormLogLevel:  logger.Info,
			SlowThreshold: defaultSlowThreshold,
			Colorful:      true,
		}
	default:
		return &platform.GormLoggerConfig{
			GormLogLevel:  logger.Warn,
			SlowThreshold: defaultSlowThreshold,
			Colorful:      true,
		}
	}
}

func (r *RawConfig) newDBConfig() *platform.DBConfig {
	return &platform.DBConfig{
		DSN: r.DSN,
	}
}

func (r *RawConfig) newGinConfig() *platform.GinConfig {
	switch r.AppEnv {
	case envDevelopment:
		return &platform.GinConfig{
			GinMode:        gin.DebugMode,
			TrustedProxies: nil,
		}
	case envProduction:
		return &platform.GinConfig{
			GinMode:        gin.ReleaseMode,
			TrustedProxies: nil,
		}
	default:
		return &platform.GinConfig{
			GinMode:        gin.DebugMode,
			TrustedProxies: nil,
		}
	}
}

func (r *RawConfig) newServerConfig() *platform.ServerConfig {
	return &platform.ServerConfig{
		HTTPAddr:          r.HTTPAddr,
		ReadHeaderTimeout: 5 * time.Second,
		ReadTimeout:       5 * time.Second,
		WriteTimeout:      5 * time.Second,
	}
}

func (r *RawConfig) newSnowflakeConfig() *platform.SnowflakeConfig {
	return &platform.SnowflakeConfig{
		NodeID: r.SnowflakeID,
	}
}

func (r *RawConfig) newJWTConfig() (*jwt.Config, error) {
	if r.AccessTokenTTL <= 0 {
		return nil, ErrAccessTTLNotPositive
	}

	//nolint:mnd // 32 bytes is the minimum key length for HMAC-SHA256 (HS256)
	if len(r.JWTSecret) < 32 {
		return nil, ErrJWTSecretTooShort
	}

	return &jwt.Config{
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
