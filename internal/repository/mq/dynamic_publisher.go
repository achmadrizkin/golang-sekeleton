// Package mq implements interfaces.MessagePublisher as a "dynamic"
// publisher: one pkg/messaging.Publisher client per logical topic key,
// each pointed at whichever broker (Kafka, RabbitMQ, or Pub/Sub) that
// topic's config.ResolvedTopic.Type selects. A single service instance can
// therefore publish user.created to RabbitMQ and audit.logs to Kafka at
// the same time, with zero code differences — only config differs.
package mq

import (
	"context"
	"fmt"

	"go.uber.org/zap"

	"github.com/fauzie/golang-sekeleton/internal/config"
	"github.com/fauzie/golang-sekeleton/internal/repository"
	"github.com/fauzie/golang-sekeleton/pkg/logger"
	"github.com/fauzie/golang-sekeleton/pkg/messaging"
	msgkafka "github.com/fauzie/golang-sekeleton/pkg/messaging/kafka"
	msgpubsub "github.com/fauzie/golang-sekeleton/pkg/messaging/pubsub"
	msgrabbitmq "github.com/fauzie/golang-sekeleton/pkg/messaging/rabbitmq"
)

// PublisherConf is the RepoFactory input for the dynamic publisher: every
// resolved publisher topic keyed by its logical name (see
// internal/config.Config.ResolvedPublishers).
type PublisherConf struct {
	Topics map[string]config.ResolvedTopic
	Logger *logger.Logger
}

// Kind implements repository.RepoFactory.
func (c *PublisherConf) Kind() repository.Kind { return repository.KindMQPublisher }

type dynamicPublisher struct {
	// clients maps logical topic key -> connected broker client.
	clients map[string]messaging.Publisher
	// topics maps logical topic key -> its resolved config (for physical name lookup).
	topics map[string]config.ResolvedTopic
	logger *logger.Logger
}

// Build implements repository.RepoFactory: connects one client per topic.
func (c *PublisherConf) Build() (interface{}, error) {
	p := &dynamicPublisher{
		clients: make(map[string]messaging.Publisher, len(c.Topics)),
		topics:  c.Topics,
		logger:  c.Logger,
	}

	for key, topic := range c.Topics {
		client, err := buildPublisherClient(topic)
		if err != nil {
			_ = p.Close()
			return nil, fmt.Errorf("mq: build publisher for topic %q: %w", key, err)
		}
		p.clients[key] = client
	}

	return p, nil
}

func buildPublisherClient(topic config.ResolvedTopic) (messaging.Publisher, error) {
	switch topic.Type {
	case config.BrokerKafka:
		return msgkafka.NewPublisher(msgkafka.PublisherConfig{
			Brokers:      topic.Kafka.Brokers,
			Topic:        topic.Name,
			Compression:  topic.Kafka.Compression,
			RequiredAcks: topic.Kafka.RequiredAcks,
		}), nil

	case config.BrokerRabbitMQ:
		return msgrabbitmq.NewPublisher(msgrabbitmq.PublisherConfig{
			Conn: msgrabbitmq.ConnConfig{
				Host:     topic.RabbitMQ.Host,
				Port:     topic.RabbitMQ.Port,
				VHost:    topic.RabbitMQ.VHost,
				Username: topic.RabbitMQ.Username,
				Password: topic.RabbitMQ.Password,
			},
			Queue:        topic.Name,
			DeliveryMode: uint8(topic.RabbitMQ.DeliveryMode),
			MessageTTL:   topic.RabbitMQ.MessageTTL,
			EnableDLQ:    topic.EnableDLQ,
			QueueType:    topic.RabbitMQ.QueueType,
		})

	case config.BrokerPubSub:
		return msgpubsub.NewPublisher(context.Background(), msgpubsub.Config{
			ProjectID:    topic.PubSub.ProjectID,
			Topic:        topic.Name,
			EmulatorHost: topic.PubSub.EmulatorHost,
		})

	default:
		return nil, fmt.Errorf("unknown broker type %q", topic.Type)
	}
}

// publish looks up topicKey's client and physical name, then publishes
// value with headers. Every typed PublishXxx method funnels through this.
func (p *dynamicPublisher) publish(ctx context.Context, topicKey string, value []byte, headers map[string]string) error {
	client, exists := p.clients[topicKey]
	if !exists {
		return fmt.Errorf("mq: no publisher configured for topic key %q", topicKey)
	}
	topicName := p.topics[topicKey].Name
	if topicName == "" {
		topicName = topicKey
	}
	if err := client.PublishWithMetadata(ctx, topicName, value, headers); err != nil {
		return err
	}
	p.logger.Info("published message",
		zap.String("topic_key", topicKey), zap.String("actual_topic", topicName), zap.Int("message_size", len(value)))
	return nil
}

func (p *dynamicPublisher) Close() error {
	var lastErr error
	for _, c := range p.clients {
		if err := c.Close(); err != nil {
			lastErr = err
		}
	}
	return lastErr
}

func (p *dynamicPublisher) Ping(ctx context.Context) error {
	for key, c := range p.clients {
		if err := c.Ping(ctx); err != nil {
			return fmt.Errorf("mq: ping topic %q: %w", key, err)
		}
	}
	return nil
}
