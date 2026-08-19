// Package endpoint sits between delivery (gRPC/HTTP/messaging) and
// usecase. Every operation is adapted to the same function shape, which is
// what lets middleware (currently: singleflight on reads) be attached
// per-operation without touching the handler or the usecase.
package endpoint

import (
	"context"

	"golang.org/x/sync/singleflight"
)

// Endpoint represents a single RPC/operation, uniform across every
// transport.
type Endpoint func(ctx context.Context, request interface{}) (response interface{}, err error)

// KeyFunc derives a singleflight dedup key from a request.
type KeyFunc func(request interface{}) string

var sfGroup singleflight.Group

// withSingleflight collapses concurrent identical requests (same key) into
// one underlying call, so a cache-cold moment doesn't turn into a thundering
// herd against the database. Only ever applied to read operations —
// collapsing two different writes into one would silently drop one of them.
func withSingleflight(next Endpoint, key KeyFunc) Endpoint {
	return func(ctx context.Context, request interface{}) (interface{}, error) {
		k := key(request)
		v, err, _ := sfGroup.Do(k, func() (interface{}, error) {
			return next(ctx, request)
		})
		return v, err
	}
}
