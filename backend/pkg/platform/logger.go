package platform

import (
	"log"
	"log/slog"
	"os"
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

// SlogConfig defines the configuration interface for the application logger.
type SlogConfig struct {
	LogLevel  slog.Level
	LogFormat SlogFormat
}

// NewLogger creates a structured slog.Logger based on the provided config.
func NewLogger(cfg *SlogConfig) *slog.Logger {
	opts := &slog.HandlerOptions{Level: cfg.LogLevel}
	var handler slog.Handler
	if cfg.LogFormat == LogFormatJSON {
		handler = slog.NewJSONHandler(os.Stdout, opts)
	} else {
		handler = slog.NewTextHandler(os.Stdout, opts)
	}
	return slog.New(handler)
}

// GormLoggerConfig defines the configuration interface for the GORM database logger.
type GormLoggerConfig struct {
	GormLogLevel  logger.LogLevel
	SlowThreshold time.Duration
	Colorful      bool
}

// NewGormLogger creates a GORM logger adapter using the provided config.
func NewGormLogger(cfg *GormLoggerConfig) logger.Interface {
	return logger.New(
		log.New(os.Stdout, "\r\n", log.LstdFlags),
		logger.Config{
			SlowThreshold:             cfg.SlowThreshold,
			LogLevel:                  cfg.GormLogLevel,
			IgnoreRecordNotFoundError: true,
			ParameterizedQueries:      true,
			Colorful:                  cfg.Colorful,
		},
	)
}
