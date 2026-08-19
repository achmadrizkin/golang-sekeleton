// Package logger wraps zap so every log line can be correlated with the
// active OpenTelemetry trace without every call site having to remember to
// extract it.
package logger

import (
	"context"

	"go.opentelemetry.io/otel/trace"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

// Logger wraps *zap.Logger. Kept as a thin type (instead of using *zap.Logger
// directly) so WithContext can live next to the rest of our logging
// conventions and so call sites don't need to import zap just to log.
type Logger struct {
	*zap.Logger
}

// Config controls logger construction.
type Config struct {
	Level       string // debug|info|warn|error
	Environment string // development|production
	Encoding    string // json|console
}

// New builds a Logger from Config. Development environments get a
// human-readable console encoder and debug-friendly defaults; anything else
// gets structured JSON suitable for log aggregation.
func New(cfg Config) (*Logger, error) {
	level := zapcore.InfoLevel
	if cfg.Level != "" {
		_ = level.UnmarshalText([]byte(cfg.Level))
	}

	var zcfg zap.Config
	if cfg.Environment == "development" {
		zcfg = zap.NewDevelopmentConfig()
	} else {
		zcfg = zap.NewProductionConfig()
	}
	zcfg.Level = zap.NewAtomicLevelAt(level)
	if cfg.Encoding != "" {
		zcfg.Encoding = cfg.Encoding
	}

	base, err := zcfg.Build(zap.AddCallerSkip(1))
	if err != nil {
		return nil, err
	}
	return &Logger{Logger: base}, nil
}

// WithContext returns a child logger with the active trace_id/span_id fields
// attached, so log lines can be correlated with traces in the observability
// backend. Safe to call with a context that carries no active span.
func (l *Logger) WithContext(ctx context.Context) *Logger {
	span := trace.SpanFromContext(ctx)
	sc := span.SpanContext()
	if !sc.IsValid() {
		return l
	}
	return &Logger{Logger: l.Logger.With(
		zap.String("trace_id", sc.TraceID().String()),
		zap.String("span_id", sc.SpanID().String()),
	)}
}

// Sync flushes any buffered log entries. Call on shutdown.
func (l *Logger) Sync() error {
	return l.Logger.Sync()
}
