// Package slogx provides a loosely coupled, translation-focused factory for constructing
// log/slog handlers, including the standard JSON/Text handlers and the third-party tint handler.
package slogx

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"os"
	"strings"
	"time"

	"github.com/lmittmann/tint"
	"go.opentelemetry.io/otel/trace"
)

var (
	errUnsupportedSlogLevel   = errors.New("unsupported slog level")
	errUnsupportedSlogHandler = errors.New("unsupported slog handler")
	errUnsupportedTimeFormat  = errors.New("unsupported time format")
)

func ParseLevel(level string) (slog.Level, error) {
	switch strings.ToLower(strings.TrimSpace(level)) {
	case "debug":
		return slog.LevelDebug, nil
	case "info":
		return slog.LevelInfo, nil
	case "warn":
		return slog.LevelWarn, nil
	case "error":
		return slog.LevelError, nil
	default:
		return slog.LevelInfo, fmt.Errorf("%w: %q", errUnsupportedSlogLevel, level)
	}
}

var slogTimeFormatMap = map[string]string{
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

func ParseTimeFormat(s string) (string, error) {
	format, ok := slogTimeFormatMap[strings.ToLower(strings.TrimSpace(s))]
	if !ok {
		return "", fmt.Errorf("%w: %q", errUnsupportedTimeFormat, s)
	}

	return format, nil
}

type TintOption func(*tint.Options)

func WithTimeFormat(format string) TintOption {
	return func(o *tint.Options) {
		o.TimeFormat = format
	}
}

func WithoutColor() TintOption {
	return func(o *tint.Options) {
		o.NoColor = true
	}
}

func NewHandler(handlerName string, slogOpts *slog.HandlerOptions, tintOpts ...TintOption) (slog.Handler, error) {
	if slogOpts == nil {
		slogOpts = &slog.HandlerOptions{}
	}

	switch strings.ToLower(strings.TrimSpace(handlerName)) {
	case "default":
		return slog.Default().Handler(), nil
	case "discard":
		return slog.DiscardHandler, nil
	case "json":
		return slog.NewJSONHandler(os.Stderr, slogOpts), nil
	case "text":
		return slog.NewTextHandler(os.Stderr, slogOpts), nil
	case "tint":
		opts := &tint.Options{
			AddSource:   slogOpts.AddSource,
			Level:       slogOpts.Level,
			ReplaceAttr: slogOpts.ReplaceAttr,
			TimeFormat:  time.StampMilli,
			NoColor:     false,
		}

		for _, opt := range tintOpts {
			opt(opts)
		}

		return tint.NewHandler(os.Stderr, opts), nil

	default:
		return nil, fmt.Errorf("%w: %q", errUnsupportedSlogHandler, handlerName)
	}
}

type OtelHandler struct {
	slog.Handler
}

func NewOtelHandler(next slog.Handler) slog.Handler {
	return &OtelHandler{
		Handler: next,
	}
}

func (h *OtelHandler) Handle(ctx context.Context, r slog.Record) error {
	span := trace.SpanFromContext(ctx)

	if span.SpanContext().IsValid() {
		r.AddAttrs(
			slog.String("trace_id", span.SpanContext().TraceID().String()),
			slog.String("span_id", span.SpanContext().SpanID().String()),
		)
	}

	return h.Handler.Handle(ctx, r)
}

func (h *OtelHandler) WithAttrs(attrs []slog.Attr) slog.Handler {
	return &OtelHandler{Handler: h.Handler.WithAttrs(attrs)}
}

func (h *OtelHandler) WithGroup(name string) slog.Handler {
	return &OtelHandler{Handler: h.Handler.WithGroup(name)}
}
