// Package telemetry wires up RITA's OpenTelemetry tracing.
//
// Tracing is opt-in: InitTracer only installs a real exporter when one of the
// standard OTEL_EXPORTER_OTLP_ENDPOINT / OTEL_EXPORTER_OTLP_TRACES_ENDPOINT
// environment variables is set. Otherwise it leaves the global no-op tracer in
// place and returns a shutdown function that does nothing, so running RITA
// without a collector has zero overhead and no configuration burden.
package telemetry

import (
	"context"
	"os"
	"time"

	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/exporters/otlp/otlptrace/otlptracehttp"
	"go.opentelemetry.io/otel/propagation"
	"go.opentelemetry.io/otel/sdk/resource"
	sdktrace "go.opentelemetry.io/otel/sdk/trace"
)

// ShutdownFunc flushes and stops the tracer provider. It is always safe to call,
// even when tracing was never enabled.
type ShutdownFunc func(context.Context) error

// InitTracer configures the global OpenTelemetry tracer provider. The OTLP/HTTP
// exporter is configured entirely from the standard OTEL_* environment
// variables. When no OTLP endpoint is configured, tracing stays disabled and the
// returned ShutdownFunc is a no-op.
func InitTracer(ctx context.Context, serviceName, version string) (ShutdownFunc, error) {
	noop := func(context.Context) error { return nil }

	if os.Getenv("OTEL_EXPORTER_OTLP_ENDPOINT") == "" && os.Getenv("OTEL_EXPORTER_OTLP_TRACES_ENDPOINT") == "" {
		return noop, nil
	}

	exp, err := otlptracehttp.New(ctx)
	if err != nil {
		return nil, err
	}

	res, err := resource.Merge(resource.Default(), resource.NewSchemaless(
		attribute.String("service.name", serviceName),
		attribute.String("service.version", version),
	))
	if err != nil {
		return nil, err
	}

	tp := sdktrace.NewTracerProvider(
		sdktrace.WithBatcher(exp),
		sdktrace.WithResource(res),
	)
	otel.SetTracerProvider(tp)
	otel.SetTextMapPropagator(propagation.NewCompositeTextMapPropagator(
		propagation.TraceContext{}, propagation.Baggage{},
	))

	return func(ctx context.Context) error {
		ctx, cancel := context.WithTimeout(ctx, 5*time.Second)
		defer cancel()
		return tp.Shutdown(ctx)
	}, nil
}
