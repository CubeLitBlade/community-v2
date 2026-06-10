package bootstrap

import (
	"errors"
	"fmt"
	"log/slog"
	"os"
	"strings"
	"time"

	"github.com/lmittmann/tint"
	"go.uber.org/fx"

	"github.com/cubelitblade/community-v2/backend/services/authz/internal/shared"
)

var (
	errUnsupportedSlogLevel      = errors.New("unsupported slog level")
	errUnsupportedTintTimeFormat = errors.New("unsupported tint time format")
	errUnsupportedSlogHandler    = errors.New("unsupported slog handler")
)

var (
	slogLevelMap = map[string]slog.Level{
		"debug": slog.LevelDebug,
		"info":  slog.LevelInfo,
		"warn":  slog.LevelWarn,
		"error": slog.LevelError,
	}

	slogTimeFormatMap = map[string]string{
		"rfc3339":      time.RFC3339,
		"rfc3339_nano": time.RFC3339Nano,
		"kitchen":      time.Kitchen,
		"ruby":         time.RubyDate,
		"layout":       time.Layout,
		"date_time":    time.DateTime,
		"ansic":        time.ANSIC,
		"unix_date":    time.UnixDate,
		"stamp":        time.Stamp,
		"stamp_milli":  time.StampMilli,
		"stamp_micro":  time.StampMicro,
		"stamp_nano":   time.StampNano,
	}
)

func slogHandler(cfg *shared.SlogConfig) (slog.Handler, error) {
	switch strings.ToLower(cfg.Type) {
	case "default":
		return slog.Default().Handler(), nil
	case "discard":
		return slog.DiscardHandler, nil
	}

	slogLevel, ok := slogLevelMap[strings.ToLower(cfg.Level)]
	if !ok {
		return nil, fmt.Errorf("%w: %q", errUnsupportedSlogLevel, cfg.Level)
	}

	switch strings.ToLower(cfg.Type) {
	case "json":
		return slog.NewJSONHandler(os.Stderr, &slog.HandlerOptions{
			Level: slogLevel,
		}), nil
	case "text":
		return slog.NewTextHandler(os.Stderr, &slog.HandlerOptions{
			Level: slogLevel,
		}), nil
	case "tint":
		format, ok := slogTimeFormatMap[strings.ToLower(cfg.Tint.TimeFormat)]
		if !ok {
			return nil, fmt.Errorf("%w: %q", errUnsupportedTintTimeFormat, cfg.Tint.TimeFormat)
		}

		return tint.NewHandler(os.Stderr, &tint.Options{
			Level:      slogLevel,
			TimeFormat: format,
		}), nil
	}

	return nil, fmt.Errorf("%w: %q", errUnsupportedSlogHandler, cfg.Type)
}

func provideSlogLogger(cfg *shared.SlogConfig) (*slog.Logger, error) {
	handler, err := slogHandler(cfg)
	if err != nil {
		return nil, fmt.Errorf("build slog handler: %w", err)
	}

	return slog.New(handler), nil
}

// SlogModule provides the slog fx module.
func SlogModule() fx.Option {
	return fx.Options(
		fx.Provide(provideSlogLogger),
	)
}
