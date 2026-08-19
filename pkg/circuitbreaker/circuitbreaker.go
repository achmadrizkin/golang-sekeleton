// Package circuitbreaker wraps sony/gobreaker so outbound gRPC/REST client
// calls can be protected without every call site re-implementing the same
// threshold/timeout bookkeeping.
package circuitbreaker

import (
	"context"
	"fmt"

	"github.com/sony/gobreaker/v2"
)

// Config controls when the breaker trips and how long it stays open.
type Config struct {
	Name                string
	MaxRequests         uint32  // requests allowed through while half-open
	IntervalSeconds     int     // closed-state counter reset interval, 0 = never
	TimeoutSeconds      int     // how long the breaker stays open before half-open
	FailureRatio        float64 // trips when failures/requests >= this ratio
	MinRequestThreshold uint32  // minimum requests before the ratio is evaluated
}

// Breaker wraps a typed gobreaker.CircuitBreaker[T] for T = any, matching
// how call sites use it: Execute a func that returns (interface{}, error).
type Breaker struct {
	cb *gobreaker.CircuitBreaker[any]
}

// New builds a Breaker from Config.
func New(cfg Config) *Breaker {
	settings := gobreaker.Settings{
		Name:        cfg.Name,
		MaxRequests: valueOr(cfg.MaxRequests, 1),
		Interval:    secondsOr(cfg.IntervalSeconds, 0),
		Timeout:     secondsOr(cfg.TimeoutSeconds, 30),
		ReadyToTrip: func(counts gobreaker.Counts) bool {
			threshold := cfg.MinRequestThreshold
			if threshold == 0 {
				threshold = 5
			}
			ratio := cfg.FailureRatio
			if ratio <= 0 {
				ratio = 0.6
			}
			return counts.Requests >= threshold &&
				float64(counts.TotalFailures)/float64(counts.Requests) >= ratio
		},
	}
	return &Breaker{cb: gobreaker.NewCircuitBreaker[any](settings)}
}

// Execute runs fn through the breaker. When the breaker is open it returns
// gobreaker.ErrOpenState immediately without calling fn.
func (b *Breaker) Execute(_ context.Context, fn func() (interface{}, error)) (interface{}, error) {
	res, err := b.cb.Execute(func() (any, error) { return fn() })
	if err != nil {
		return nil, fmt.Errorf("circuitbreaker[%s]: %w", b.cb.Name(), err)
	}
	return res, nil
}

// State reports the current breaker state (closed/open/half-open) for
// health/metrics reporting.
func (b *Breaker) State() gobreaker.State {
	return b.cb.State()
}
