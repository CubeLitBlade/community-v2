// Package logger provides constructors for the structured application logger
// and the GORM database logger.
package logger

import (
	"log"
	"log/slog"
	"os"

	"github.com/cubelitblade/community-v2/backend/pkg/platform/config"
	"gorm.io/gorm/logger"
)

// NewLogger creates a structured slog.Logger based on the provided config.
func NewLogger(cfg *config.SlogConfig) *slog.Logger {
	opts := &slog.HandlerOptions{Level: cfg.LogLevel}
	var handler slog.Handler
	if cfg.LogFormat == config.LogFormatJSON {
		handler = slog.NewJSONHandler(os.Stdout, opts)
	} else {
		handler = slog.NewTextHandler(os.Stdout, opts)
	}
	return slog.New(handler)
}

// NewGormLogger creates a GORM logger adapter using the provided config.
func NewGormLogger(cfg *config.GormLoggerConfig) logger.Interface {
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
