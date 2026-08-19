// Package messaging holds one consumer per topic. A consumer's job is
// narrow: reconstruct claims/trace context from the message headers, call
// the same endpoint the HTTP/gRPC paths use, and translate the result into
// an ack (return nil) or a retry (return an error) — never business logic.
package messaging

import (
	"context"
	"encoding/json"

	"go.uber.org/zap"

	"github.com/fauzie/golang-sekeleton/internal/endpoint"
	"github.com/fauzie/golang-sekeleton/pkg/claims"
	apperrors "github.com/fauzie/golang-sekeleton/pkg/errors"
	"github.com/fauzie/golang-sekeleton/pkg/logger"
	"github.com/fauzie/golang-sekeleton/pkg/telemetry"
)

// UserCreatedMessage is the wire shape published to user.created.
type UserCreatedMessage struct {
	ID string `json:"id"`
}

// UserCreatedConsumer handles user.created messages.
type UserCreatedConsumer struct {
	Logger   *logger.Logger
	Endpoint endpoint.Endpoint
}

// NewUserCreatedConsumer builds a UserCreatedConsumer.
func NewUserCreatedConsumer(log *logger.Logger, ep endpoint.Endpoint) *UserCreatedConsumer {
	return &UserCreatedConsumer{Logger: log, Endpoint: ep}
}

// Handle processes one user.created delivery. Matches
// pkg/messaging.Handler's signature so it can be passed straight to
// Consumer.Start.
func (c *UserCreatedConsumer) Handle(ctx context.Context, topic, subscription string, headers map[string]string, body []byte) error {
	ctx = claims.SetClaims(ctx, claims.FromMetadata(headers))

	ctx, end := telemetry.StartMessagingTransaction(ctx, topic, headers)
	defer end()

	log := c.Logger.WithContext(ctx)

	var msg UserCreatedMessage
	if err := json.Unmarshal(body, &msg); err != nil {
		log.Error("failed to unmarshal message", zap.Error(err))
		telemetry.CaptureError(ctx, err)
		return nil // ACK: malformed JSON will not improve on retry
	}

	span, ctx := telemetry.StartSpan(ctx, "process_user_created", "business_logic")
	defer span.End()

	resp, err := c.Endpoint(ctx, endpoint.ProcessUserCreatedRequest{ID: msg.ID})
	if err != nil {
		log.Error("failed to process user.created event", zap.Error(err))
		if !apperrors.IsRetryable(err) {
			return nil // ACK: terminal error, do not retry forever
		}
		return err // NACK -> redelivered per the topic's DeliveryPolicy
	}

	if response, ok := resp.(endpoint.ProcessUserCreatedResponse); ok && response.Err != nil {
		if !apperrors.IsRetryable(response.Err) {
			return nil
		}
		return response.Err
	}

	telemetry.SetTransactionResult(ctx, "success")
	return nil
}
