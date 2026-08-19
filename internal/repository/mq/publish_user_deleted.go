package mq

import (
	"context"
	"encoding/json"

	"github.com/fauzie/golang-sekeleton/internal/domain"
	"github.com/fauzie/golang-sekeleton/pkg/claims"
	apperrors "github.com/fauzie/golang-sekeleton/pkg/errors"
	"github.com/fauzie/golang-sekeleton/pkg/telemetry"
)

// PublishUserDeleted publishes a user.deleted event.
func (p *dynamicPublisher) PublishUserDeleted(ctx context.Context, payload *domain.UserDeletedPayload) error {
	span, ctx := telemetry.StartSpan(ctx, "publish_user_deleted", "messaging")
	defer span.End()

	body, err := json.Marshal(payload)
	if err != nil {
		telemetry.CaptureError(ctx, err)
		return apperrors.NewDataAccessError("failed to marshal user.deleted message", err)
	}

	headers := map[string]string{}
	telemetry.InjectMessagingContext(ctx, headers)
	for k, v := range claims.GetClaims(ctx).ToMetadata() {
		headers[k] = v
	}

	if err := p.publish(ctx, TopicUserDeleted, body, headers); err != nil {
		telemetry.CaptureError(ctx, err)
		telemetry.SetLabel(ctx, "error", true)
		return apperrors.NewDataAccessError("failed to publish user.deleted event", err)
	}
	return nil
}

// [INJECTION POINT: Publisher Methods]
