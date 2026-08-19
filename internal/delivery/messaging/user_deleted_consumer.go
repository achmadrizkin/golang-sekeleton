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

// UserDeletedMessage is the wire shape published to user.deleted.
type UserDeletedMessage struct {
	ID string `json:"id"`
}

// UserDeletedConsumer handles user.deleted messages.
type UserDeletedConsumer struct {
	Logger   *logger.Logger
	Endpoint endpoint.Endpoint
}

// NewUserDeletedConsumer builds a UserDeletedConsumer.
func NewUserDeletedConsumer(log *logger.Logger, ep endpoint.Endpoint) *UserDeletedConsumer {
	return &UserDeletedConsumer{Logger: log, Endpoint: ep}
}

// Handle processes one user.deleted delivery.
func (c *UserDeletedConsumer) Handle(ctx context.Context, topic, subscription string, headers map[string]string, body []byte) error {
	ctx = claims.SetClaims(ctx, claims.FromMetadata(headers))

	ctx, end := telemetry.StartMessagingTransaction(ctx, topic, headers)
	defer end()

	log := c.Logger.WithContext(ctx)

	var msg UserDeletedMessage
	if err := json.Unmarshal(body, &msg); err != nil {
		log.Error("failed to unmarshal message", zap.Error(err))
		telemetry.CaptureError(ctx, err)
		return nil
	}

	span, ctx := telemetry.StartSpan(ctx, "process_user_deleted", "business_logic")
	defer span.End()

	resp, err := c.Endpoint(ctx, endpoint.ProcessUserDeletedRequest{ID: msg.ID})
	if err != nil {
		log.Error("failed to process user.deleted event", zap.Error(err))
		if !apperrors.IsRetryable(err) {
			return nil
		}
		return err
	}

	if response, ok := resp.(endpoint.ProcessUserDeletedResponse); ok && response.Err != nil {
		if !apperrors.IsRetryable(response.Err) {
			return nil
		}
		return response.Err
	}

	telemetry.SetTransactionResult(ctx, "success")
	return nil
}
