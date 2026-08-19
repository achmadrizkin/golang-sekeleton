// Package telemetry wires OpenTelemetry tracing. HTTP, gRPC, SQL, and Redis
// get automatic instrumentation from their respective otel* packages at the
// call sites in internal/server and internal/repository; this package
// covers everything else: process-wide setup, manual spans, and carrying
// trace context across the message-broker boundary (which has no built-in
// otel instrumentation the way HTTP/gRPC do).
package telemetry

import (
	"context"
	"fmt"

	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/codes"
	"go.opentelemetry.io/otel/exporters/otlp/otlptrace/otlptracegrpc"
	"go.opentelemetry.io/otel/propagation"
	sdkresource "go.opentelemetry.io/otel/sdk/resource"
	sdktrace "go.opentelemetry.io/otel/sdk/trace"
	semconv "go.opentelemetry.io/otel/semconv/v1.26.0"
	"go.opentelemetry.io/otel/trace"
)

// Config controls telemetry initialization.
type Config struct {
	Enabled        bool
	ServiceName    string
	ServiceVersion string
	Environment    string
	OTLPEndpoint   string // host:port, e.g. "otel-collector:4317"
	SampleRatio    float64
}

var tracer trace.Tracer = otel.Tracer("noop")

// Init configures the global TracerProvider and text-map propagator. It
// returns a shutdown func to call during graceful shutdown, and is a no-op
// (dependency-free noop tracer) when cfg.Enabled is false so telemetry can
// be disabled in local dev without touching call sites.
func Init(ctx context.Context, cfg Config) (func(context.Context) error, error) {
	otel.SetTextMapPropagator(propagation.NewCompositeTextMapPropagator(
		propagation.TraceContext{}, propagation.Baggage{},
	))

	if !cfg.Enabled {
		return func(context.Context) error { return nil }, nil
	}

	res, err := sdkresource.New(ctx,
		sdkresource.WithAttributes(
			semconv.ServiceName(cfg.ServiceName),
			semconv.ServiceVersion(cfg.ServiceVersion),
			semconv.DeploymentEnvironment(cfg.Environment),
		),
	)
	if err != nil {
		return nil, fmt.Errorf("telemetry: build resource: %w", err)
	}

	exporter, err := otlptracegrpc.New(ctx,
		otlptracegrpc.WithEndpoint(cfg.OTLPEndpoint),
		otlptracegrpc.WithInsecure(),
	)
	if err != nil {
		return nil, fmt.Errorf("telemetry: build OTLP exporter: %w", err)
	}

	ratio := cfg.SampleRatio
	if ratio <= 0 {
		ratio = 1
	}

	tp := sdktrace.NewTracerProvider(
		sdktrace.WithBatcher(exporter),
		sdktrace.WithResource(res),
		sdktrace.WithSampler(sdktrace.ParentBased(sdktrace.TraceIDRatioBased(ratio))),
	)
	otel.SetTracerProvider(tp)
	tracer = tp.Tracer(cfg.ServiceName)

	return tp.Shutdown, nil
}

// StartSpan starts a child span named name, tagged with a "component"
// attribute (e.g. "business_logic", "messaging"), and returns it together
// with the context carrying it.
func StartSpan(ctx context.Context, name, component string) (trace.Span, context.Context) {
	ctx, span := tracer.Start(ctx, name, trace.WithAttributes(attribute.String("component", component)))
	return span, ctx
}

// CaptureError records err on the span active in ctx and marks it as an
// error, without altering control flow.
func CaptureError(ctx context.Context, err error) {
	if err == nil {
		return
	}
	span := trace.SpanFromContext(ctx)
	span.RecordError(err)
	span.SetStatus(codes.Error, err.Error())
}

// SetLabel attaches a key/value attribute to the span active in ctx.
func SetLabel(ctx context.Context, key string, value interface{}) {
	span := trace.SpanFromContext(ctx)
	switch v := value.(type) {
	case string:
		span.SetAttributes(attribute.String(key, v))
	case bool:
		span.SetAttributes(attribute.Bool(key, v))
	case int:
		span.SetAttributes(attribute.Int(key, v))
	default:
		span.SetAttributes(attribute.String(key, fmt.Sprintf("%v", v)))
	}
}

// TraceID returns the active trace ID in ctx, or "" if none.
func TraceID(ctx context.Context) string {
	sc := trace.SpanContextFromContext(ctx)
	if !sc.IsValid() {
		return ""
	}
	return sc.TraceID().String()
}
