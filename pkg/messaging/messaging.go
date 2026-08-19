// Package messaging defines the broker-agnostic contracts used everywhere
// else in the codebase. internal/repository/mq builds one Publisher and,
// where configured, one Consumer per logical topic key, picking the
// concrete adapter (kafka, rabbitmq, pubsub — see the sibling packages) from
// config. Nothing above this package (usecase, endpoint, delivery) ever
// imports a broker-specific package directly.
package messaging

import "context"

// Publisher sends messages to one physical topic/queue on one broker.
type Publisher interface {
	// Publish sends value with no extra headers.
	Publish(ctx context.Context, topic string, value []byte) error
	// PublishWithMetadata sends value with headers (used to propagate trace
	// context and JWT claims across the broker boundary).
	PublishWithMetadata(ctx context.Context, topic string, value []byte, headers map[string]string) error
	Close() error
	Ping(ctx context.Context) error
}

// Handler processes one delivered message. Returning nil acknowledges the
// message (it will not be redelivered); returning an error triggers the
// broker adapter's retry/backoff/DLQ policy (DeliveryPolicy) before the
// message is finally acked or dead-lettered. This mirrors delivery
// semantics: the *caller* (internal/delivery/messaging) decides whether an
// error is retryable via pkg/errors.IsRetryable and returns nil for
// terminal errors so they are not retried forever.
type Handler func(ctx context.Context, topic, subscription string, headers map[string]string, body []byte) error

// Consumer subscribes to one physical topic/subscription on one broker and
// invokes Handler for every delivered message until Stop is called or ctx
// is cancelled.
type Consumer interface {
	Start(ctx context.Context, handler Handler) error
	Stop() error
}

// DeliveryPolicy configures redelivery behaviour shared by every broker
// adapter's consumer loop: how many times to retry in-process, the backoff
// curve between attempts, and whether to shunt the message to a dead-letter
// topic after exhausting retries instead of dropping it silently.
type DeliveryPolicy struct {
	ConcurrentConsumers int
	ExponentialBackoff  bool
	InitialInterval     int // seconds
	MaxInterval         int // seconds
	Multiplier          float64
	MaxRetries          int
	EnableDLQ           bool
}

// DefaultDeliveryPolicy is used when a topic does not override any retry
// settings.
func DefaultDeliveryPolicy() DeliveryPolicy {
	return DeliveryPolicy{
		ConcurrentConsumers: 1,
		ExponentialBackoff:  true,
		InitialInterval:     2,
		MaxInterval:         30,
		Multiplier:          2,
		MaxRetries:          5,
		EnableDLQ:           true,
	}
}
