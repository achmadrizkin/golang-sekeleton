package server

import (
	"context"
	"fmt"

	"go.uber.org/zap"

	"github.com/fauzie/golang-sekeleton/internal/config"
	deliverymessaging "github.com/fauzie/golang-sekeleton/internal/delivery/messaging"
	"github.com/fauzie/golang-sekeleton/internal/endpoint"
	"github.com/fauzie/golang-sekeleton/internal/repository/mq"
	"github.com/fauzie/golang-sekeleton/pkg/logger"
	"github.com/fauzie/golang-sekeleton/pkg/messaging"
)

// startConsumers builds and starts one broker consumer per configured
// consumer topic whose logical key has a registered handler below. A
// topic key with no case here only logs a warning at startup — matching
// the doc's documented failure mode ("No consumer handler implementation
// for topic key") instead of crashing the whole service over one
// misconfigured topic.
func startConsumers(ctx context.Context, cfg *config.Config, log *logger.Logger, userEndpoints endpoint.UserEndpoints) ([]messaging.Consumer, error) {
	var started []messaging.Consumer

	for key, topic := range cfg.ResolvedConsumers() {
		var handler messaging.Handler

		// [INJECTION POINT: Consumer Handler Mapping]
		switch key {
		case mq.TopicUserCreated:
			handler = deliverymessaging.NewUserCreatedConsumer(log, userEndpoints.ProcessUserCreatedEndpoint).Handle
		case mq.TopicUserDeleted:
			handler = deliverymessaging.NewUserDeletedConsumer(log, userEndpoints.ProcessUserDeletedEndpoint).Handle
		default:
			log.Warn("no consumer handler implementation for topic key", zap.String("key", key))
			continue
		}

		consumer, err := mq.BuildConsumer(topic)
		if err != nil {
			stopAll(started)
			return nil, fmt.Errorf("server: build consumer for topic %q: %w", key, err)
		}
		if err := consumer.Start(ctx, handler); err != nil {
			stopAll(started)
			return nil, fmt.Errorf("server: start consumer for topic %q: %w", key, err)
		}
		started = append(started, consumer)
		log.Info("consumer started", zap.String("topic_key", key), zap.String("physical_topic", topic.Name), zap.String("broker", string(topic.Type)))
	}

	return started, nil
}

func stopAll(consumers []messaging.Consumer) {
	for _, c := range consumers {
		_ = c.Stop()
	}
}
