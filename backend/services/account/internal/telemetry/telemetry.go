package telemetry

import (
	"go.opentelemetry.io/otel/metric"
	"go.opentelemetry.io/otel/trace"
)

type Telemetry struct {
	mp metric.MeterProvider
	tp trace.TracerProvider
}

func NewTelemetry(mp metric.MeterProvider, tp trace.TracerProvider) *Telemetry {
	return &Telemetry{mp: mp, tp: tp}
}

func (t *Telemetry) Meter(name string) metric.Meter {
	return t.mp.Meter(name)
}

func (t *Telemetry) Tracer(name string) trace.Tracer {
	return t.tp.Tracer(name)
}
