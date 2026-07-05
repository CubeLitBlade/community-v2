package telemetry

import (
	"go.opentelemetry.io/otel/metric"
	"go.opentelemetry.io/otel/trace"
)

var cacheDurationBuckets = []float64{0.0001, 0.0005, 0.001, 0.005, 0.01, 0.05, 0.1, 0.5, 1.0}

type Telemetry struct {
	MeterProvider  metric.MeterProvider
	TracerProvider trace.TracerProvider
}

func (t *Telemetry) Meter(name string, options ...metric.MeterOption) metric.Meter {
	return t.MeterProvider.Meter(name, options...)
}

func (t *Telemetry) Tracer(name string, options ...trace.TracerOption) trace.Tracer {
	return t.TracerProvider.Tracer(name, options...)
}
