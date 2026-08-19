// Package retry provides a small fixed-delay retry helper for outbound
// client calls (REST/gRPC clients to other services). Message-broker
// redelivery has its own policy — see pkg/messaging.WithRetry — this
// package is specifically for synchronous outbound calls.
package retry

import (
	"context"
	"time"
)

// Config controls attempt count and delay between attempts.
type Config struct {
	MaxAttempts int
	Delay       time.Duration
}

// Do runs fn until it succeeds, ctx is cancelled, or cfg.MaxAttempts is
// reached (minimum 1). Returns the last error on exhaustion.
func Do(ctx context.Context, cfg Config, fn func() error) error {
	attempts := cfg.MaxAttempts
	if attempts < 1 {
		attempts = 1
	}
	delay := cfg.Delay
	if delay <= 0 {
		delay = 200 * time.Millisecond
	}

	var lastErr error
	for i := 0; i < attempts; i++ {
		if err := ctx.Err(); err != nil {
			return err
		}
		if lastErr = fn(); lastErr == nil {
			return nil
		}
		if i == attempts-1 {
			break
		}
		select {
		case <-ctx.Done():
			return ctx.Err()
		case <-time.After(delay):
		}
	}
	return lastErr
}
