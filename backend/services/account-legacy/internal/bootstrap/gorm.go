package bootstrap

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"log/slog"
	"strings"
	"time"

	"go.uber.org/fx"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"

	"github.com/cubelitblade/community-v2/backend/pkg/platform"
	"github.com/cubelitblade/community-v2/backend/services/account-legacy/internal/shared"
)

var errUnsupportedGormLogLevel = errors.New("unsupported gorm log level")

var gormLoggerLevelMap = map[string]logger.LogLevel{
	"silent": logger.Silent,
	"info":   logger.Info,
	"warn":   logger.Warn,
	"error":  logger.Error,
}

func provideGorm(
	dbCfg *shared.DatabaseConfig, gormCfg *shared.GormConfig, s *slog.Logger, lc fx.Lifecycle,
) (gormDB *gorm.DB, sqlDB *sql.DB, err error) {
	g, err := gormConfig(gormCfg, s)
	if err != nil {
		return nil, nil, fmt.Errorf("gorm config: %w", err)
	}

	gormDB, sqlDB, err = platform.Open(dbCfg.URL, g)
	if err != nil {
		return nil, nil, fmt.Errorf("open database: %w", err)
	}

	lc.Append(fx.Hook{
		OnStop: func(ctx context.Context) error {
			return sqlDB.Close()
		},
	})

	return gormDB, sqlDB, nil
}

func gormConfig(cfg *shared.GormConfig, s *slog.Logger) (*gorm.Config, error) {
	slogGormLogger, err := gormLogger(cfg, s)
	if err != nil {
		return nil, fmt.Errorf("create gorm logger: %w", err)
	}

	return &gorm.Config{
		SkipDefaultTransaction:                   cfg.SkipDefaultTransaction,
		PrepareStmt:                              cfg.PrepareStmt,
		DisableForeignKeyConstraintWhenMigrating: cfg.DisableForeignKeyConstraintWhenMigrating,
		Logger:                                   slogGormLogger,
	}, nil
}

func gormLogger(cfg *shared.GormConfig, slogLogger *slog.Logger) (logger.Interface, error) {
	level, ok := gormLoggerLevelMap[strings.ToLower(cfg.Logger.Level)]
	if !ok {
		return nil, fmt.Errorf("%w: %q", errUnsupportedGormLogLevel, cfg.Logger.Level)
	}

	return NewSlogGormLogger(slogLogger, logger.Config{
		SlowThreshold:             cfg.Logger.SlowThreshold,
		LogLevel:                  level,
		IgnoreRecordNotFoundError: cfg.Logger.IgnoreRecordNotFoundError,
		ParameterizedQueries:      cfg.Logger.ParameterizedQuery,
		Colorful:                  cfg.Logger.Colorful,
	}), nil
}

// SlogGormLogger implements the gorm logger.Interface using slog.
type SlogGormLogger struct {
	slogLogger                *slog.Logger
	level                     logger.LogLevel
	slowThreshold             time.Duration
	ignoreRecordNotFoundError bool
}

// NewSlogGormLogger creates a new SlogGormLogger instance with the provided slog.Logger and gorm logger.Config.
func NewSlogGormLogger(slogLogger *slog.Logger, cfg logger.Config) *SlogGormLogger {
	return &SlogGormLogger{
		slogLogger:                slogLogger.With(slog.String("module", "gorm")),
		level:                     cfg.LogLevel,
		slowThreshold:             cfg.SlowThreshold,
		ignoreRecordNotFoundError: cfg.IgnoreRecordNotFoundError,
	}
}

// LogMode changes the logger level and returns a new instance of the logger.
func (l *SlogGormLogger) LogMode(level logger.LogLevel) logger.Interface {
	newLogger := *l
	newLogger.level = level

	return &newLogger
}

// Info logs a message at the Info level.
func (l *SlogGormLogger) Info(ctx context.Context, format string, args ...any) {
	if l.level >= logger.Info {
		l.slogLogger.InfoContext(ctx, "gorm info", slog.String("msg", fmt.Sprintf(format, args...)))
	}
}

// Warn logs a message at the Warn level.
func (l *SlogGormLogger) Warn(ctx context.Context, format string, args ...any) {
	if l.level >= logger.Warn {
		l.slogLogger.WarnContext(ctx, "gorm warning", slog.String("msg", fmt.Sprintf(format, args...)))
	}
}

// Error logs a message at the Error level.
func (l *SlogGormLogger) Error(ctx context.Context, format string, args ...any) {
	if l.level >= logger.Error {
		l.slogLogger.ErrorContext(ctx, "gorm error", slog.String("msg", fmt.Sprintf(format, args...)))
	}
}

// Trace logs SQL execution details, handling slow queries and errors.
func (l *SlogGormLogger) Trace(
	ctx context.Context, begin time.Time, fc func() (statement string, rowsAffected int64), err error,
) {
	if l.level <= logger.Silent {
		return
	}

	elapsed := time.Since(begin)
	statement, rows := fc()

	attrs := []any{
		slog.String("sql", statement),
		slog.Int64("rows", rows),
		slog.Duration("elapsed", elapsed),
	}

	if err != nil && (!l.ignoreRecordNotFoundError || !errors.Is(err, gorm.ErrRecordNotFound)) {
		l.slogLogger.ErrorContext(ctx, "GORM query error", append(attrs, slog.String("err", err.Error()))...)
		return
	}

	if l.slowThreshold > 0 && elapsed > l.slowThreshold {
		l.slogLogger.WarnContext(ctx, "GORM slow query", append(attrs, slog.Duration("threshold", l.slowThreshold))...)
		return
	}

	if l.level >= logger.Info {
		l.slogLogger.InfoContext(ctx, "GORM query", attrs...)
	}
}

// GormModule provides the database connection fx module.
func GormModule() fx.Option {
	return fx.Options(
		fx.Provide(provideGorm),
	)
}
