package mq

import (
	"context"
	"encoding/json"

	"github.com/fauzie/golang-sekeleton/internal/domain"
	"github.com/fauzie/golang-sekeleton/pkg/claims"
	apperrors "github.com/fauzie/golang-sekeleton/pkg/errors"
	"github.com/fauzie/golang-sekeleton/pkg/telemetry"
)

// PublishUserCreated publishes a user.created event. Trace context and the
// caller's claims are propagated via message headers so a consumer on the
// other side of the broker can continue the same trace and know who
// triggered the event.
func (p *dynamicPublisher) PublishUserCreated(ctx context.Context, payload *domain.UserCreatedPayload) error {
	span, ctx := telemetry.StartSpan(ctx, "publish_user_created", "messaging")
	defer span.End()

	body, err := json.Marshal(payload)
	if err != nil {
		telemetry.CaptureError(ctx, err)
		return apperrors.NewDataAccessError("failed to marshal user.created message", err)
	}

	headers := map[string]string{}
	telemetry.InjectMessagingContext(ctx, headers)
	for k, v := range claims.GetClaims(ctx).ToMetadata() {
		headers[k] = v
	}

	if err := p.publish(ctx, TopicUserCreated, body, headers); err != nil {
		telemetry.CaptureError(ctx, err)
		telemetry.SetLabel(ctx, "error", true)
		return apperrors.NewDataAccessError("failed to publish user.created event", err)
	}
	return nil
}
