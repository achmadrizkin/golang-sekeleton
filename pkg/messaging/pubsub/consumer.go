package pubsub

import (
	"context"
	"fmt"

	"cloud.google.com/go/pubsub"

	"github.com/fauzie/golang-sekeleton/pkg/messaging"
)

// ConsumerConfig configures a Pub/Sub subscription consumer.
type ConsumerConfig struct {
	ProjectID    string
	Topic        string
	Subscription string
	EmulatorHost string
	Policy       messaging.DeliveryPolicy
}

type consumer struct {
	cfg    ConsumerConfig
	client *pubsub.Client
	sub    *pubsub.Subscription
	dlqPub messaging.Publisher
	cancel context.CancelFunc
	done   chan struct{}
}

// NewConsumer connects and ensures cfg.Subscription exists, creating it
// bound to cfg.Topic if necessary (dev convenience, same rationale as
// ensureTopic in pubsub.go). When cfg.Policy.EnableDLQ is set it also
// prepares a publisher for the "<topic>-dlq" topic used once in-process
// retries are exhausted.
func NewConsumer(ctx context.Context, cfg ConsumerConfig) (messaging.Consumer, error) {
	client, err := newClient(ctx, cfg.ProjectID, cfg.EmulatorHost)
	if err != nil {
		return nil, err
	}

	topic, err := ensureTopic(ctx, client, cfg.Topic)
	if err != nil {
		_ = client.Close()
		return nil, err
	}

	sub := client.Subscription(cfg.Subscription)
	ok, err := sub.Exists(ctx)
	if err != nil {
		_ = client.Close()
		return nil, fmt.Errorf("pubsub: check subscription %q: %w", cfg.Subscription, err)
	}
	if !ok {
		if sub, err = client.CreateSubscription(ctx, cfg.Subscription, pubsub.SubscriptionConfig{Topic: topic}); err != nil {
			_ = client.Close()
			return nil, fmt.Errorf("pubsub: create subscription %q: %w", cfg.Subscription, err)
		}
	}

	concurrency := cfg.Policy.ConcurrentConsumers
	if concurrency < 1 {
		concurrency = 1
	}
	sub.ReceiveSettings.NumGoroutines = concurrency

	var dlqPub messaging.Publisher
	if cfg.Policy.EnableDLQ {
		dlqPub, err = NewPublisher(ctx, Config{ProjectID: cfg.ProjectID, Topic: cfg.Topic + "-dlq", EmulatorHost: cfg.EmulatorHost})
		if err != nil {
			_ = client.Close()
			return nil, err
		}
	}

	return &consumer{cfg: cfg, client: client, sub: sub, dlqPub: dlqPub, done: make(chan struct{})}, nil
}

func (c *consumer) Start(ctx context.Context, handler messaging.Handler) error {
	ctx, cancel := context.WithCancel(ctx)
	c.cancel = cancel

	go func() {
		defer close(c.done)
		_ = c.sub.Receive(ctx, func(msgCtx context.Context, m *pubsub.Message) {
			err := messaging.WithRetry(msgCtx, c.cfg.Policy, func() error {
				return handler(msgCtx, c.cfg.Topic, c.cfg.Subscription, m.Attributes, m.Data)
			})
			if err == nil {
				m.Ack()
				return
			}

			if c.dlqPub != nil {
				if pubErr := c.dlqPub.PublishWithMetadata(msgCtx, c.cfg.Topic+"-dlq", m.Data, m.Attributes); pubErr == nil {
					m.Ack()
					return
				}
			}
			// Nack: Pub/Sub redelivers according to the subscription's own
			// (native) retry/backoff policy.
			m.Nack()
		})
	}()
	return nil
}

func (c *consumer) Stop() error {
	if c.cancel != nil {
		c.cancel()
	}
	<-c.done
	if c.dlqPub != nil {
		_ = c.dlqPub.Close()
	}
	return c.client.Close()
}
