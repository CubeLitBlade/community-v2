// Package shared defines the shared types and interfaces used across the account service.
package shared

import (
	"time"
)

// Config represents the top-level configuration for the account service.
type Config struct {
	ServiceName string `koanf:"service_name"`
	HTTPAddr    string `koanf:"http_addr"`

	Database  *DatabaseConfig  `koanf:"database"`
	Cookie    *CookieConfig    `koanf:"cookie"`
	Redis     *RedisConfig     `koanf:"redis"`
	RabbitMQ  *RabbitMQConfig  `koanf:"rabbitmq"`
	JWT       *JWTConfig       `koanf:"jwt"`
	Snowflake *SnowflakeConfig `koanf:"snowflake"`
	Token     *TokenConfig     `koanf:"token"`
	Slog      *SlogConfig      `koanf:"slog"`
	Gorm      *GormConfig      `koanf:"gorm"`
	Gin       *GinConfig       `koanf:"gin"`
}

// DatabaseConfig holds the configuration for the database connection.
type DatabaseConfig struct {
	URL string `koanf:"url"`
}

// CookieConfig holds the configuration for HTTP cookies.
type CookieConfig struct {
	Secure   bool   `koanf:"secure"`
	SameSite string `koanf:"same_site"`
}

// RedisConfig holds the configuration for the Redis client connection.
type RedisConfig struct {
	Addr string `koanf:"addr"`
}

// RabbitMQConfig holds the configuration for the RabbitMQ message broker connection.
type RabbitMQConfig struct {
	BrokerURL string `koanf:"broker_url"`
}

// JWTConfig holds the configuration for JSON Web Token authentication.
type JWTConfig struct {
	Secret string `koanf:"secret"`
}

// SnowflakeConfig holds the configuration for the Snowflake distributed ID generator.
type SnowflakeConfig struct {
	NodeID int64 `koanf:"node_id"`
}

// TokenConfig holds the configuration for token issuance and validity periods.
type TokenConfig struct {
	TokenIssuer         string        `koanf:"issuer"`
	AccessTokenValidity time.Duration `koanf:"access_token_validity"`
}

// SlogConfig holds the configuration for the slog structured logger.
type SlogConfig struct {
	Level string         `koanf:"level"`
	Type  string         `koanf:"type"`
	Tint  SlogTintConfig `koanf:"tint"`
}

// SlogTintConfig holds the configuration specific to the tint slog handler.
type SlogTintConfig struct {
	TimeFormat string `koanf:"time_format"`
}

// GormConfig holds the configuration for the GORM.
type GormConfig struct {
	SkipDefaultTransaction                   bool              `koanf:"skip_default_transaction"`
	PrepareStmt                              bool              `koanf:"prepare_stmt"`
	DisableForeignKeyConstraintWhenMigrating bool              `koanf:"disable_foreign_key_constraint_when_migrating"`
	Logger                                   *GormLoggerConfig `koanf:"logger"`
}

// GormLoggerConfig holds the configuration for the GORM logger.
type GormLoggerConfig struct {
	Level                     string        `koanf:"level"`
	SlowThreshold             time.Duration `koanf:"slow_threshold"`
	IgnoreRecordNotFoundError bool          `koanf:"ignore_record_not_found_error"`
	ParameterizedQuery        bool          `koanf:"parameterized_query"`
	Colorful                  bool          `koanf:"colorful"`
}

// GinConfig holds the configuration for the Gin web framework.
type GinConfig struct {
	Mode string `koanf:"mode"`
}
