package bootstrap

import (
	"fmt"
	"net/http"
	"os"
	"strconv"
	"time"
)

const (
	envAddr                  = "ADDR"
	envDatabaseURL           = "DATABASE_URL"
	envSnowflakeID           = "SNOWFLAKE_ID"
	envJWTSecret             = "JWT_SECRET"
	envAccessTokenTTL        = "JWT_ACCESS_TOKEN_TTL"
	envCookieSecure          = "COOKIE_SECURE"
	envAccessTokenCookieName = "ACCESS_TOKEN_COOKIE_NAME"
)

const (
	defaultAddr                  = ":8080"
	defaultSnowflakeID           = int64(1)
	defaultAccessTokenTTL        = 2 * time.Hour
	defaultCookieSecure          = false
	defaultAccessTokenCookieName = "ACCESS_TOKEN"
	defaultCookieSameSite        = http.SameSiteLaxMode
)

const (
	minJWTSecretBytes = 32

	envIntParseBase = 10
	envInt64Bits    = 64
)

var (
	errDatabaseURLUnset = fmt.Errorf(
		"%s unset",
		envDatabaseURL,
	)
	errJWTSecretTooShort = fmt.Errorf(
		"%s should be at least %d bytes",
		envJWTSecret,
		minJWTSecretBytes,
	)
	errAccessTokenTTLPositive = fmt.Errorf(
		"%s must be positive",
		envAccessTokenTTL,
	)
	errAccessTokenCookieNameEmpty = fmt.Errorf(
		"%s must not be empty",
		envAccessTokenCookieName,
	)
)

// Config holds the application configuration values.
type Config struct {
	Addr                  string
	DatabaseURL           string
	AccessTokenCookieName string
	JWTSecret             string
	SnowflakeID           int64
	AccessTokenTTL        time.Duration
	CookieSameSite        http.SameSite
	CookieSecure          bool
}

// LoadConfig reads configuration from environment variables and validates them.
func LoadConfig() (Config, error) {
	snowflakeID, err := envInt64(envSnowflakeID, defaultSnowflakeID)
	if err != nil {
		return Config{}, err
	}

	accessTokenTTL, err := envDuration(envAccessTokenTTL, defaultAccessTokenTTL)
	if err != nil {
		return Config{}, err
	}

	cookieSecure, err := envBool(envCookieSecure, defaultCookieSecure)
	if err != nil {
		return Config{}, err
	}

	cfg := Config{
		Addr:        envString(envAddr, defaultAddr),
		DatabaseURL: os.Getenv(envDatabaseURL),

		SnowflakeID: snowflakeID,

		JWTSecret:      os.Getenv(envJWTSecret),
		AccessTokenTTL: accessTokenTTL,

		CookieSecure:   cookieSecure,
		CookieSameSite: defaultCookieSameSite,
		AccessTokenCookieName: envString(
			envAccessTokenCookieName,
			defaultAccessTokenCookieName,
		),
	}

	if err := cfg.validate(); err != nil {
		return Config{}, err
	}

	return cfg, nil
}

func (c *Config) validate() error {
	if c.DatabaseURL == "" {
		return errDatabaseURLUnset
	}

	if len(c.JWTSecret) < minJWTSecretBytes {
		return errJWTSecretTooShort
	}

	if c.AccessTokenTTL <= 0 {
		return errAccessTokenTTLPositive
	}

	if c.AccessTokenCookieName == "" {
		return errAccessTokenCookieNameEmpty
	}

	return nil
}

func envString(key, fallback string) string {
	value := os.Getenv(key)
	if value == "" {
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

	parsed, err := strconv.ParseInt(value, envIntParseBase, envInt64Bits)
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
