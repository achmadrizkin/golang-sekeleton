package rabbitmq

import (
	"context"
	"fmt"
	"sync"

	amqp "github.com/rabbitmq/amqp091-go"

	"github.com/fauzie/golang-sekeleton/pkg/messaging"
)

// ConsumerConfig configures a consumer for one queue ("Subscription").
// Topic is the logical name reported to Handler; Subscription is the queue
// actually consumed from (defaults to Topic when empty), matching the
// doc's separation between logical topic key and physical queue name.
type ConsumerConfig struct {
	Conn          ConnConfig
	Topic         string
	Subscription  string
	PrefetchCount int
	Policy        messaging.DeliveryPolicy
}

type consumer struct {
	cfg  ConsumerConfig
	conn *amqp.Connection
	chs  []*amqp.Channel
	wg   sync.WaitGroup
}

// NewConsumer connects and prepares (but does not yet start) a consumer.
func NewConsumer(cfg ConsumerConfig) (messaging.Consumer, error) {
	if cfg.Subscription == "" {
		cfg.Subscription = cfg.Topic
	}
	conn, err := amqp.Dial(cfg.Conn.URL())
	if err != nil {
		return nil, fmt.Errorf("rabbitmq: dial: %w", err)
	}
	return &consumer{cfg: cfg, conn: conn}, nil
}

func (c *consumer) Start(ctx context.Context, handler messaging.Handler) error {
	concurrency := c.cfg.Policy.ConcurrentConsumers
	if concurrency < 1 {
		concurrency = 1
	}
	prefetch := c.cfg.PrefetchCount
	if prefetch <= 0 {
		prefetch = 10
	}

	for i := 0; i < concurrency; i++ {
		ch, err := c.conn.Channel()
		if err != nil {
			return fmt.Errorf("rabbitmq: open channel: %w", err)
		}
		if err := ch.Qos(prefetch, 0, false); err != nil {
			return fmt.Errorf("rabbitmq: set qos: %w", err)
		}
		if _, err := ch.QueueDeclarePassive(c.cfg.Subscription, true, false, false, false, nil); err != nil {
			// Queue may not exist yet if the publisher side hasn't run;
			// declare it defensively (durable, no DLX — publisher owns DLX
			// wiring since it knows EnableDLQ policy at creation time).
			if _, derr := ch.QueueDeclare(c.cfg.Subscription, true, false, false, false, nil); derr != nil {
				return fmt.Errorf("rabbitmq: declare queue %q: %w", c.cfg.Subscription, derr)
			}
		}

		deliveries, err := ch.Consume(c.cfg.Subscription, "", false, false, false, false, nil)
		if err != nil {
			return fmt.Errorf("rabbitmq: consume %q: %w", c.cfg.Subscription, err)
		}

		c.chs = append(c.chs, ch)
		c.wg.Add(1)
		go c.consumeLoop(ctx, deliveries, handler)
	}
	return nil
}

func (c *consumer) consumeLoop(ctx context.Context, deliveries <-chan amqp.Delivery, handler messaging.Handler) {
	defer c.wg.Done()

	for {
		select {
		case <-ctx.Done():
			return
		case d, ok := <-deliveries:
			if !ok {
				return
			}

			headers := make(map[string]string, len(d.Headers))
			for k, v := range d.Headers {
				if s, ok := v.(string); ok {
					headers[k] = s
				} else {
					headers[k] = fmt.Sprintf("%v", v)
				}
			}

			err := messaging.WithRetry(ctx, c.cfg.Policy, func() error {
				return handler(ctx, c.cfg.Topic, c.cfg.Subscription, headers, d.Body)
			})

			if err == nil {
				_ = d.Ack(false)
				continue
			}

			// Exhausted in-process retries. requeue=false: if the queue was
			// declared with a dead-letter-exchange (EnableDLQ at publish
			// time), the broker routes it to the DLQ automatically;
			// otherwise the message is dropped rather than requeued
			// immediately, which would otherwise tight-loop on a poison
			// message.
			_ = d.Nack(false, false)
		}
	}
}

func (c *consumer) Stop() error {
	for _, ch := range c.chs {
		_ = ch.Close()
	}
	c.wg.Wait()
	return c.conn.Close()
}
