// Package config provides configuration types for the platform infrastructure
// (database, HTTP server, Gin engine, logging, Snowflake ID generation).
package config

import (
	"log/slog"
	"time"

	"gorm.io/gorm/logger"
)

// SlogFormat is the output format for the structured logger.
type SlogFormat string

const (
	// LogFormatJSON outputs log entries as JSON.
	LogFormatJSON SlogFormat = "json"
	// LogFormatText outputs log entries as plain text.
	LogFormatText SlogFormat = "text"
)

// SlogConfig defines the configuration for the application logger.
type SlogConfig struct {
	LogLevel  slog.Level
	LogFormat SlogFormat
}

// GormLoggerConfig defines the configuration for the GORM database logger.
type GormLoggerConfig struct {
	GormLogLevel  logger.LogLevel
	SlowThreshold time.Duration
	Colorful      bool
}

// DBConfig provides the Data Source Name for the database connection.
type DBConfig struct {
	DSN string
}

// ServerConfig holds the HTTP server settings.
type ServerConfig struct {
	HTTPAddr          string
	ReadHeaderTimeout time.Duration
	ReadTimeout       time.Duration
	WriteTimeout      time.Duration
}

// GinConfig provides settings for the Gin engine.
type GinConfig struct {
	GinMode        string
	TrustedProxies []string
}

// SnowflakeConfig holds the worker node ID for Snowflake ID generation.
type SnowflakeConfig struct {
	NodeID int64
}
