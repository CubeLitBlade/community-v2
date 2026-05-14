package bootstrap

import (
	"fmt"
	"net/http"
	"os"
	"strconv"
	"time"
)

type Config struct {
	Addr                  string
	DatabaseURL           string
	SnowflakeID           int64
	JWTSecret             []byte
	AccessTokenTTL        time.Duration
	CSRFAuthKey           []byte
	CookieSecure          bool
	CookieSameSite        http.SameSite
	AccessTokenCookieName string
}

func LoadConfig() (Config, error) {
	snowflakeID, err := envInt64("SNOWFLAKE_ID", 1)
	if err != nil {
		return Config{}, err
	}

	accessTokenTTL, err := envDuration("JWT_ACCESS_TOKEN_TTL", 2*time.Hour)
	if err != nil {
		return Config{}, err
	}

	cookieSecure, err := envBool("COOKIE_SECURE", false)
	if err != nil {
		return Config{}, err
	}

	cfg := Config{
		Addr:        envString("ADDR", ":8080"),
		DatabaseURL: os.Getenv("DATABASE_URL"),

		SnowflakeID: snowflakeID,

		JWTSecret:      []byte(os.Getenv("JWT_SECRET")),
		AccessTokenTTL: accessTokenTTL,

		CSRFAuthKey: []byte(os.Getenv("CSRF_AUTH_KEY")),

		CookieSecure:          cookieSecure,
		CookieSameSite:        http.SameSiteLaxMode,
		AccessTokenCookieName: envString("ACCESS_TOKEN_COOKIE_NAME", "ACCESS_TOKEN"),
	}

	if err := cfg.validate(); err != nil {
		return Config{}, err
	}

	return cfg, nil
}

func (c Config) validate() error {
	if c.DatabaseURL == "" {
		return fmt.Errorf("DATABASE_URL is unset")
	}

	if len(c.JWTSecret) < 32 {
		return fmt.Errorf("JWT_SECRET must be at least 32 bytes")
	}

	if c.AccessTokenTTL <= 0 {
		return fmt.Errorf("JWT_ACCESS_TOKEN_TTL must be positive")
	}

	if len(c.CSRFAuthKey) != 32 {
		return fmt.Errorf("CSRF_AUTH_KEY must be exactly 32 bytes")
	}

	if c.AccessTokenCookieName == "" {
		return fmt.Errorf("ACCESS_TOKEN_COOKIE_NAME must not be empty")
	}

	return nil
}

func envString(key string, fallback string) string {
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
		return false, fmt.Errorf("parse %s: %w", key, err)
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
		return 0, fmt.Errorf("parse %s: %w", key, err)
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
		return 0, fmt.Errorf("parse %s: %w", key, err)
	}

	return parsed, nil
}
