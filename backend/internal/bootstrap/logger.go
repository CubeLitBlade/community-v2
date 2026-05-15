package bootstrap

import (
	"log"
	"log/slog"
	"os"
	"time"

	"gorm.io/gorm/logger"
)

const (
	defaultSlowThreshold = 200 * time.Millisecond
)

// NewAppLogger creates and returns a new slog.Logger for application-wide
// logging.
func NewAppLogger() *slog.Logger {
	return slog.New(slog.NewTextHandler(os.Stdout, &slog.HandlerOptions{
		Level: slog.LevelDebug,
	}))
}

// NewGormLogger creates and returns a new GORM logger.Interface for database
// operation logging.
func NewGormLogger() logger.Interface {
	return logger.New(
		log.New(os.Stdout, "\r\n", log.LstdFlags),
		logger.Config{
			SlowThreshold:             defaultSlowThreshold,
			LogLevel:                  logger.Warn,
			IgnoreRecordNotFoundError: true,
			ParameterizedQueries:      false,
			Colorful:                  true,
		},
	)
}
