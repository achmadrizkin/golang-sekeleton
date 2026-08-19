// Package kafka implements pkg/messaging.Publisher and Consumer on top of
// segmentio/kafka-go — a pure-Go client (no cgo/librdkafka dependency),
// which keeps cross-platform builds (including this Windows dev box) simple.
package kafka

import (
	"context"
	"crypto/tls"
	"fmt"
	"time"

	kafkago "github.com/segmentio/kafka-go"

	"github.com/fauzie/golang-sekeleton/pkg/messaging"
)

// PublisherConfig configures a Kafka producer for one physical topic.
type PublisherConfig struct {
	Brokers      []string
	Topic        string
	ClientID     string
	RequiredAcks int    // -1 = all, 0 = none, 1 = leader
	Compression  string // none|gzip|snappy|lz4|zstd
	BatchTimeout time.Duration
	TLSEnabled   bool
}

type publisher struct {
	writer *kafkago.Writer
}

// NewPublisher builds a Kafka producer bound to cfg.Topic.
func NewPublisher(cfg PublisherConfig) messaging.Publisher {
	w := &kafkago.Writer{
		Addr:         kafkago.TCP(cfg.Brokers...),
		Topic:        cfg.Topic,
		Balancer:     &kafkago.LeastBytes{},
		RequiredAcks: kafkago.RequiredAcks(cfg.RequiredAcks),
		Compression:  parseCompression(cfg.Compression),
		BatchTimeout: nonZero(cfg.BatchTimeout, 10*time.Millisecond),
	}
	if cfg.TLSEnabled {
		w.Transport = &kafkago.Transport{TLS: &tls.Config{MinVersion: tls.VersionTLS12}}
	}
	return &publisher{writer: w}
}

func (p *publisher) Publish(ctx context.Context, topic string, value []byte) error {
	return p.PublishWithMetadata(ctx, topic, value, nil)
}

func (p *publisher) PublishWithMetadata(ctx context.Context, topic string, value []byte, headers map[string]string) error {
	msg := kafkago.Message{
		Topic: topic,
		Value: value,
		Time:  time.Now(),
	}
	for k, v := range headers {
		msg.Headers = append(msg.Headers, kafkago.Header{Key: k, Value: []byte(v)})
	}
	if err := p.writer.WriteMessages(ctx, msg); err != nil {
		return fmt.Errorf("kafka: publish to %q: %w", topic, err)
	}
	return nil
}

func (p *publisher) Close() error { return p.writer.Close() }

func (p *publisher) Ping(ctx context.Context) error {
	conn, err := kafkago.DialContext(ctx, "tcp", p.writer.Addr.String())
	if err != nil {
		return fmt.Errorf("kafka: ping: %w", err)
	}
	defer func() { _ = conn.Close() }()
	return nil
}

func parseCompression(c string) kafkago.Compression {
	switch c {
	case "gzip":
		return kafkago.Gzip
	case "snappy":
		return kafkago.Snappy
	case "lz4":
		return kafkago.Lz4
	case "zstd":
		return kafkago.Zstd
	default:
		return 0
	}
}

func nonZero(d, fallback time.Duration) time.Duration {
	if d <= 0 {
		return fallback
	}
	return d
}
