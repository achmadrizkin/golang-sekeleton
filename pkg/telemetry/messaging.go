package telemetry

import (
	"context"

	"go.opentelemetry.io/otel"
)

// headerCarrier adapts a map[string]string to propagation.TextMapCarrier so
// the standard W3C trace-context propagator can read/write message-broker
// headers exactly like it would HTTP headers.
type headerCarrier map[string]string

func (h headerCarrier) Get(key string) string { return h[key] }
func (h headerCarrier) Set(key, value string) { h[key] = value }
func (h headerCarrier) Keys() []string {
	keys := make([]string, 0, len(h))
	for k := range h {
		keys = append(keys, k)
	}
	return keys
}

// InjectMessagingContext writes the trace context active in ctx into
// headers, so a consumer on the other side of the broker can continue the
// same trace via StartMessagingTransaction.
func InjectMessagingContext(ctx context.Context, headers map[string]string) {
	otel.GetTextMapPropagator().Inject(ctx, headerCarrier(headers))
}

// StartMessagingTransaction extracts any trace context carried in headers,
// starts a span named "consume "+topic linked to the producer's trace, and
// returns the resulting context plus an end func to defer.
func StartMessagingTransaction(ctx context.Context, topic string, headers map[string]string) (context.Context, func()) {
	ctx = otel.GetTextMapPropagator().Extract(ctx, headerCarrier(headers))
	span, ctx := StartSpan(ctx, "consume "+topic, "messaging")
	return ctx, func() { span.End() }
}

// SetTransactionResult tags the active messaging transaction span with its
// terminal outcome (e.g. "success", "retry", "dropped").
func SetTransactionResult(ctx context.Context, result string) {
	SetLabel(ctx, "transaction.result", result)
}
