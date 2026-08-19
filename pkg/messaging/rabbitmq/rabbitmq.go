// Package rabbitmq implements pkg/messaging.Publisher and Consumer on top of
// amqp091-go. Each topic maps to one durable queue bound to the default
// exchange (routing key == queue name), which keeps the mapping between a
// logical topic and a physical queue obvious without requiring callers to
// think about exchange topology.
package rabbitmq

import (
	"context"
	"fmt"
	"time"

	amqp "github.com/rabbitmq/amqp091-go"

	"github.com/fauzie/golang-sekeleton/pkg/messaging"
)

// ConnConfig describes how to reach the broker.
type ConnConfig struct {
	Host     string
	Port     int
	VHost    string
	Username string
	Password string
}

// URL builds the amqp:// DSN from ConnConfig.
func (c ConnConfig) URL() string {
	vhost := c.VHost
	if vhost == "/" {
		vhost = ""
	}
	return fmt.Sprintf("amqp://%s:%s@%s:%d/%s", c.Username, c.Password, c.Host, c.Port, vhost)
}

// PublisherConfig configures a publisher for one queue.
type PublisherConfig struct {
	Conn         ConnConfig
	Queue        string
	DeliveryMode uint8 // amqp.Persistent (2) or amqp.Transient (1)
	MessageTTL   int   // seconds, 0 = no TTL
	EnableDLQ    bool
	QueueType    string // classic|quorum
}

type publisher struct {
	conn  *amqp.Connection
	ch    *amqp.Channel
	queue string
	mode  uint8
}

// NewPublisher connects, declares the queue (with DLX args when
// cfg.EnableDLQ), and returns a ready-to-use Publisher.
func NewPublisher(cfg PublisherConfig) (messaging.Publisher, error) {
	conn, err := amqp.Dial(cfg.Conn.URL())
	if err != nil {
		return nil, fmt.Errorf("rabbitmq: dial: %w", err)
	}
	ch, err := conn.Channel()
	if err != nil {
		_ = conn.Close()
		return nil, fmt.Errorf("rabbitmq: open channel: %w", err)
	}

	if err := declareQueue(ch, cfg.Queue, cfg.MessageTTL, cfg.EnableDLQ, cfg.QueueType); err != nil {
		_ = ch.Close()
		_ = conn.Close()
		return nil, err
	}

	mode := cfg.DeliveryMode
	if mode == 0 {
		mode = amqp.Persistent
	}

	return &publisher{conn: conn, ch: ch, queue: cfg.Queue, mode: mode}, nil
}

func declareQueue(ch *amqp.Channel, name string, ttlSeconds int, enableDLQ bool, queueType string) error {
	args := amqp.Table{}
	if queueType == "quorum" {
		args["x-queue-type"] = "quorum"
	}
	if ttlSeconds > 0 {
		args["x-message-ttl"] = int32(ttlSeconds * 1000)
	}
	if enableDLQ {
		dlx := name + ".dlx"
		dlq := name + ".dlq"
		args["x-dead-letter-exchange"] = dlx
		args["x-dead-letter-routing-key"] = dlq

		if err := ch.ExchangeDeclare(dlx, "direct", true, false, false, false, nil); err != nil {
			return fmt.Errorf("rabbitmq: declare DLX %q: %w", dlx, err)
		}
		if _, err := ch.QueueDeclare(dlq, true, false, false, false, nil); err != nil {
			return fmt.Errorf("rabbitmq: declare DLQ %q: %w", dlq, err)
		}
		if err := ch.QueueBind(dlq, dlq, dlx, false, nil); err != nil {
			return fmt.Errorf("rabbitmq: bind DLQ %q: %w", dlq, err)
		}
	}

	if _, err := ch.QueueDeclare(name, true, false, false, false, args); err != nil {
		return fmt.Errorf("rabbitmq: declare queue %q: %w", name, err)
	}
	return nil
}

func (p *publisher) Publish(ctx context.Context, topic string, value []byte) error {
	return p.PublishWithMetadata(ctx, topic, value, nil)
}

func (p *publisher) PublishWithMetadata(ctx context.Context, topic string, value []byte, headers map[string]string) error {
	table := amqp.Table{}
	for k, v := range headers {
		table[k] = v
	}
	err := p.ch.PublishWithContext(ctx, "", topic, false, false, amqp.Publishing{
		ContentType:  "application/json",
		DeliveryMode: p.mode,
		Timestamp:    time.Now(),
		Headers:      table,
		Body:         value,
	})
	if err != nil {
		return fmt.Errorf("rabbitmq: publish to %q: %w", topic, err)
	}
	return nil
}

func (p *publisher) Close() error {
	_ = p.ch.Close()
	return p.conn.Close()
}

func (p *publisher) Ping(_ context.Context) error {
	if p.conn == nil || p.conn.IsClosed() {
		return fmt.Errorf("rabbitmq: connection closed")
	}
	return nil
}
