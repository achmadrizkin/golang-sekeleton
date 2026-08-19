package mq

import (
	"context"
	"fmt"

	"github.com/fauzie/golang-sekeleton/internal/config"
	"github.com/fauzie/golang-sekeleton/pkg/messaging"
	msgkafka "github.com/fauzie/golang-sekeleton/pkg/messaging/kafka"
	msgpubsub "github.com/fauzie/golang-sekeleton/pkg/messaging/pubsub"
	msgrabbitmq "github.com/fauzie/golang-sekeleton/pkg/messaging/rabbitmq"
)

// BuildConsumer builds the broker-specific messaging.Consumer for one
// resolved topic. Unlike publishers, consumers are not part of Repository
// (they don't serve requests, they drive them) — internal/server builds
// one per configured consumer topic, maps it to a delivery handler by
// logical key, and owns its Start/Stop lifecycle directly.
func BuildConsumer(topic config.ResolvedTopic) (messaging.Consumer, error) {
	policy := messaging.DeliveryPolicy{
		ConcurrentConsumers: topic.ConcurrentConsumers,
		ExponentialBackoff:  topic.ExponentialBackoff,
		InitialInterval:     topic.InitialInterval,
		MaxInterval:         topic.MaxInterval,
		Multiplier:          topic.Multiplier,
		MaxRetries:          topic.MaxRetries,
		EnableDLQ:           topic.EnableDLQ,
	}

	switch topic.Type {
	case config.BrokerKafka:
		return msgkafka.NewConsumer(msgkafka.ConsumerConfig{
			Brokers: topic.Kafka.Brokers,
			Topic:   topic.Name,
			GroupID: topic.Subscription,
			Policy:  policy,
		}), nil

	case config.BrokerRabbitMQ:
		return msgrabbitmq.NewConsumer(msgrabbitmq.ConsumerConfig{
			Conn: msgrabbitmq.ConnConfig{
				Host:     topic.RabbitMQ.Host,
				Port:     topic.RabbitMQ.Port,
				VHost:    topic.RabbitMQ.VHost,
				Username: topic.RabbitMQ.Username,
				Password: topic.RabbitMQ.Password,
			},
			Topic:         topic.Name,
			Subscription:  topic.Subscription,
			PrefetchCount: topic.RabbitMQ.PrefetchCount,
			Policy:        policy,
		})

	case config.BrokerPubSub:
		return msgpubsub.NewConsumer(context.Background(), msgpubsub.ConsumerConfig{
			ProjectID:    topic.PubSub.ProjectID,
			Topic:        topic.Name,
			Subscription: topic.Subscription,
			EmulatorHost: topic.PubSub.EmulatorHost,
			Policy:       policy,
		})

	default:
		return nil, fmt.Errorf("mq: unknown broker type %q for topic %q", topic.Type, topic.Key)
	}
}
