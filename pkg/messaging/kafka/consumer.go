package kafka

import (
	"context"
	"errors"
	"sync"

	kafkago "github.com/segmentio/kafka-go"

	"github.com/fauzie/golang-sekeleton/pkg/messaging"
)

// ConsumerConfig configures a Kafka consumer group for one physical topic.
type ConsumerConfig struct {
	Brokers  []string
	Topic    string
	GroupID  string // used as the "subscription" passed to Handler
	MinBytes int
	MaxBytes int
	Policy   messaging.DeliveryPolicy
	// DLQTopic overrides the default "<topic>.dlq" dead-letter topic name.
	DLQTopic string
}

type consumer struct {
	cfg    ConsumerConfig
	dlq    *kafkago.Writer
	stopCh chan struct{}
	wg     sync.WaitGroup
}

// NewConsumer builds a Kafka consumer-group consumer bound to cfg.Topic.
func NewConsumer(cfg ConsumerConfig) messaging.Consumer {
	if cfg.MinBytes <= 0 {
		cfg.MinBytes = 1
	}
	if cfg.MaxBytes <= 0 {
		cfg.MaxBytes = 10e6
	}
	dlqTopic := cfg.DLQTopic
	if dlqTopic == "" {
		dlqTopic = cfg.Topic + ".dlq"
	}
	return &consumer{
		cfg:    cfg,
		dlq:    &kafkago.Writer{Addr: kafkago.TCP(cfg.Brokers...), Topic: dlqTopic, Balancer: &kafkago.LeastBytes{}},
		stopCh: make(chan struct{}),
	}
}

func (c *consumer) Start(ctx context.Context, handler messaging.Handler) error {
	concurrency := c.cfg.Policy.ConcurrentConsumers
	if concurrency < 1 {
		concurrency = 1
	}

	for i := 0; i < concurrency; i++ {
		reader := kafkago.NewReader(kafkago.ReaderConfig{
			Brokers:  c.cfg.Brokers,
			Topic:    c.cfg.Topic,
			GroupID:  c.cfg.GroupID,
			MinBytes: c.cfg.MinBytes,
			MaxBytes: c.cfg.MaxBytes,
		})

		c.wg.Add(1)
		go c.consumeLoop(ctx, reader, handler)
	}
	return nil
}

func (c *consumer) consumeLoop(ctx context.Context, reader *kafkago.Reader, handler messaging.Handler) {
	defer c.wg.Done()
	defer func() { _ = reader.Close() }()

	for {
		select {
		case <-ctx.Done():
			return
		case <-c.stopCh:
			return
		default:
		}

		msg, err := reader.FetchMessage(ctx)
		if err != nil {
			if errors.Is(err, context.Canceled) {
				return
			}
			continue
		}

		headers := make(map[string]string, len(msg.Headers))
		for _, h := range msg.Headers {
			headers[h.Key] = string(h.Value)
		}

		handleErr := messaging.WithRetry(ctx, c.cfg.Policy, func() error {
			return handler(ctx, c.cfg.Topic, c.cfg.GroupID, headers, msg.Value)
		})

		if handleErr == nil {
			_ = reader.CommitMessages(ctx, msg)
			continue
		}

		if c.cfg.Policy.EnableDLQ {
			dlqMsg := kafkago.Message{Value: msg.Value}
			for k, v := range headers {
				dlqMsg.Headers = append(dlqMsg.Headers, kafkago.Header{Key: k, Value: []byte(v)})
			}
			dlqMsg.Headers = append(dlqMsg.Headers, kafkago.Header{Key: "x-error", Value: []byte(handleErr.Error())})
			if pubErr := c.dlq.WriteMessages(ctx, dlqMsg); pubErr == nil {
				_ = reader.CommitMessages(ctx, msg)
			}
			// If publishing to the DLQ itself fails, leave the offset
			// uncommitted so the message is redelivered after restart.
			continue
		}
		// No DLQ: leave the offset uncommitted. The message is redelivered
		// to the group on the next restart/rebalance (at-least-once).
	}
}

func (c *consumer) Stop() error {
	close(c.stopCh)
	c.wg.Wait()
	return c.dlq.Close()
}
