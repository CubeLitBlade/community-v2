package bootstrap

import (
	"errors"
	"fmt"
	"log"
	"net/http"
	"os"
	"strconv"
	"strings"
	"time"
)

// Environment represents the deployment environment of the application.
type Environment string

// ErrInvalidConfig indicates that the application configuration is invalid.
var ErrInvalidConfig = errors.New("invalid configuration")

const (
	// EnvDevelopment represents the development environment.
	EnvDevelopment Environment = "dev"

	// EnvProduction represents the production environment.
	EnvProduction Environment = "prod"
)

// Config holds the application configuration settings.
type Config struct {
	// Env specifies the current deployment environment.
	Env Environment

	// HTTPAddr specifies the HTTP server listen address.
	HTTPAddr string

	// DatabaseURL specifies the connection string for the database.
	DatabaseURL string

	// AccessTokenCookieName specifies the name of the cookie used to store the access token.
	AccessTokenCookieName string

	// JWTSecret specifies the secret key used to sign JWT tokens.
	JWTSecret string

	// SnowflakeID specifies the machine or node ID for Snowflake ID generation.
	SnowflakeID int64

	// AccessTokenTTL specifies the time-to-live duration for access tokens.
	AccessTokenTTL time.Duration

	// CookieSameSite specifies the SameSite attribute for cookies.
	CookieSameSite http.SameSite

	// CookieSecure specifies whether cookies should be restricted to HTTPS.
	CookieSecure bool
}

// LoadConfig reads configuration from environment variables, applies defaults based on the environment,
// and validates the result. It returns a pointer to the configured Config or an error if validation fails.
func LoadConfig() (*Config, error) {
	env := Environment(envString("APP_ENV", string(EnvDevelopment)))

	var cfg Config
	cfg.Env = env

	var defaultDBURL string
	var defaultJWTSecret string
	defaultCookieSecure := false

	if env == EnvDevelopment {
		defaultDBURL = "postgres://community:community_dev_password@localhost:5432/community?sslmode=disable&search_path=account_service"
		defaultJWTSecret = "development-insecure-jwt-secret-key!!"
	}

	if env == EnvProduction {
		defaultCookieSecure = true
	}

	cfg.DatabaseURL = envString("DATABASE_URL", defaultDBURL)
	cfg.JWTSecret = envString("JWT_SECRET", defaultJWTSecret)
	cfg.HTTPAddr = envString("HTTP_ADDR", ":8080")
	cfg.AccessTokenCookieName = envString("ACCESS_TOKEN_COOKIE_NAME", "ACCESS_TOKEN")

	var err error
	if cfg.AccessTokenTTL, err = envDuration("JWT_ACCESS_TOKEN_TTL", 2*time.Hour); err != nil {
		return nil, err
	}
	if cfg.SnowflakeID, err = envInt64("SNOWFLAKE_ID", 1); err != nil {
		return nil, err
	}
	if cfg.CookieSecure, err = envBool("COOKIE_SECURE", defaultCookieSecure); err != nil {
		return nil, err
	}
	if cfg.CookieSameSite, err = envSameSite("COOKIE_SAME_SITE", http.SameSiteLaxMode); err != nil {
		return nil, err
	}

	if err := cfg.validate(); err != nil {
		return nil, err
	}

	if env == EnvDevelopment && cfg.JWTSecret == defaultJWTSecret {
		log.Println("WARNING: Using insecure default JWT_SECRET in development mode.")
	}

	return &cfg, nil
}

func (c *Config) validate() error {
	if c.AccessTokenTTL <= 0 {
		return fmt.Errorf("JWT_ACCESS_TOKEN_TTL must be positive: %w", ErrInvalidConfig)
	}
	if c.AccessTokenCookieName == "" {
		return fmt.Errorf("ACCESS_TOKEN_COOKIE_NAME must not be empty: %w", ErrInvalidConfig)
	}

	if c.Env == EnvProduction {
		if c.DatabaseURL == "" {
			return fmt.Errorf("DATABASE_URL is required in production: %w", ErrInvalidConfig)
		}

		//nolint:mnd // 32 is the standard minimum key length for HMAC-SHA256 (HS256) in JWT, context is clear in the error message
		if len(c.JWTSecret) < 32 {
			return fmt.Errorf("JWT_SECRET must be at least 32 bytes in production: %w", ErrInvalidConfig)
		}
		if !c.CookieSecure {
			return fmt.Errorf("COOKIE_SECURE must be true in production: %w", ErrInvalidConfig)
		}
	}

	return nil
}

func envString(key, fallback string) string {
	value := os.Getenv(key)
	if value == "" {
		log.Printf("%s is empty, using default value: %s\n",
			key, fallback)
		return fallback
	}

	return value
}

func envBool(key string, fallback bool) (bool, error) {
	value := os.Getenv(key)
	if value == "" {
		return fallback, nil
	}

	parsed, err := strconv.ParseBool(value)
	if err != nil {
		return false, fmt.Errorf("parse bool %s: %w", key, err)
	}

	return parsed, nil
}

func envInt64(key string, fallback int64) (int64, error) {
	value := os.Getenv(key)
	if value == "" {
		return fallback, nil
	}

	parsed, err := strconv.ParseInt(value, 10, 64)
	if err != nil {
		return 0, fmt.Errorf("parse int %s: %w", key, err)
	}

	return parsed, nil
}

func envDuration(key string, fallback time.Duration) (time.Duration, error) {
	value := os.Getenv(key)
	if value == "" {
		return fallback, nil
	}

	parsed, err := time.ParseDuration(value)
	if err != nil {
		return 0, fmt.Errorf("parse duration %s: %w", key, err)
	}

	return parsed, nil
}

func envSameSite(key string, fallback http.SameSite) (http.SameSite, error) {
	value := os.Getenv(key)
	if value == "" {
		return fallback, nil
	}

	switch strings.ToLower(value) {
	case "none":
		return http.SameSiteNoneMode, nil
	case "lax":
		return http.SameSiteLaxMode, nil
	case "strict":
		return http.SameSiteStrictMode, nil
	default:
		return 0, fmt.Errorf("invalid value %q for %s (must be strict, lax, or none): %w",
			value, key, ErrInvalidConfig)
	}
}
