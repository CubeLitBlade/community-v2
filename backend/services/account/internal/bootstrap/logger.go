package bootstrap

import (
	"log/slog"
	"os"
	"time"

	"github.com/lmittmann/tint"
	"go.uber.org/fx"
)

func slogHandler(appRT AppRuntime) slog.Handler {
	if appRT.Environment() == AppEnvProduction {
		return slog.NewJSONHandler(os.Stderr, &slog.HandlerOptions{
			Level: slog.LevelInfo,
		})
	}

	return tint.NewHandler(os.Stderr, &tint.Options{
		Level:      slog.LevelDebug,
		TimeFormat: time.StampMilli,
	})
}

func provideSlogLogger(appRT AppRuntime) *slog.Logger {
	return slog.New(slogHandler(appRT))
}

// LoggerModule provides the structured logger fx module.
func LoggerModule() fx.Option {
	return fx.Options(
		fx.Provide(provideSlogLogger),
	)
}
