package messaging

import (
	"context"
	"time"
)

// WithRetry runs fn until it succeeds, ctx is cancelled, or policy.MaxRetries
// in-process attempts are exhausted (0 or negative MaxRetries means "try
// once, no retry"). Sleeps between attempts follow an exponential (or fixed,
// when policy.ExponentialBackoff is false) backoff clamped to
// policy.MaxInterval. It returns the last error from fn, or nil on success.
//
// This is intentionally broker-agnostic: it is the shared building block
// every adapter's consumer loop (kafka/rabbitmq/pubsub) uses before falling
// back to its own DLQ behaviour.
func WithRetry(ctx context.Context, policy DeliveryPolicy, fn func() error) error {
	interval := time.Duration(policy.InitialInterval) * time.Second
	if interval <= 0 {
		interval = 2 * time.Second
	}
	maxInterval := time.Duration(policy.MaxInterval) * time.Second
	if maxInterval <= 0 {
		maxInterval = 30 * time.Second
	}
	multiplier := policy.Multiplier
	if multiplier <= 1 {
		multiplier = 2
	}

	attempts := policy.MaxRetries + 1
	if attempts < 1 {
		attempts = 1
	}

	var lastErr error
	for i := 0; i < attempts; i++ {
		if err := ctx.Err(); err != nil {
			return err
		}

		lastErr = fn()
		if lastErr == nil {
			return nil
		}

		if i == attempts-1 {
			break
		}

		select {
		case <-ctx.Done():
			return ctx.Err()
		case <-time.After(interval):
		}

		if policy.ExponentialBackoff {
			interval = time.Duration(float64(interval) * multiplier)
			if interval > maxInterval {
				interval = maxInterval
			}
		}
	}
	return lastErr
}
