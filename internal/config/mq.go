package config

// BrokerType selects which pkg/messaging adapter a topic uses.
type BrokerType string

const (
	BrokerKafka    BrokerType = "kafka"
	BrokerRabbitMQ BrokerType = "rabbitmq"
	BrokerPubSub   BrokerType = "pubsub"
)

// KafkaConfig is the Kafka-specific slice of a broker config block.
type KafkaConfig struct {
	Brokers      []string `mapstructure:"brokers"`
	Compression  string   `mapstructure:"compression"`
	RequiredAcks int      `mapstructure:"required_acks"`
}

// RabbitMQConfig is the RabbitMQ-specific slice of a broker config block.
type RabbitMQConfig struct {
	Host          string `mapstructure:"host"`
	Port          int    `mapstructure:"port"`
	VHost         string `mapstructure:"vhost"`
	Username      string `mapstructure:"username"`
	Password      string `mapstructure:"password"`
	MessageTTL    int    `mapstructure:"message_ttl"`
	DeliveryMode  int    `mapstructure:"delivery_mode"`
	QueueType     string `mapstructure:"queue_type"` // classic|quorum
	PrefetchCount int    `mapstructure:"prefetch_count"`
}

// PubSubConfig is the Pub/Sub-specific slice of a broker config block.
type PubSubConfig struct {
	ProjectID    string `mapstructure:"project_id"`
	EmulatorHost string `mapstructure:"emulator_host"`
}

// GeneralMQConfig holds the default settings for each broker type, applied
// to every topic of that type unless overridden per-topic.
type GeneralMQConfig struct {
	ConcurrentConsumers int            `mapstructure:"concurrent_consumers"`
	ExponentialBackoff  bool           `mapstructure:"exponential_backoff"`
	InitialInterval     int            `mapstructure:"initial_interval"`
	MaxInterval         int            `mapstructure:"max_interval"`
	Multiplier          float64        `mapstructure:"multiplier"`
	MaxRetries          int            `mapstructure:"max_retries"`
	EnableDLQ           bool           `mapstructure:"enable_dlq"`
	Kafka               KafkaConfig    `mapstructure:"kafka"`
	RabbitMQ            RabbitMQConfig `mapstructure:"rabbitmq"`
	PubSub              PubSubConfig   `mapstructure:"pubsub"`
}

// TopicOverride carries only the keys a topic wants to override from
// general_mq_config. All fields are optional/zero-value-means-unset except
// where noted; the merge in Resolve treats an explicitly-zero override as
// "not set" (Viper cannot easily distinguish "absent" from "zero" for
// scalars in a nested map, so this mirrors the doc's own limitation and is
// documented in README).
type TopicOverride struct {
	ConcurrentConsumers int     `mapstructure:"concurrent_consumers"`
	ExponentialBackoff  *bool   `mapstructure:"exponential_backoff"`
	InitialInterval     int     `mapstructure:"initial_interval"`
	MaxInterval         int     `mapstructure:"max_interval"`
	Multiplier          float64 `mapstructure:"multiplier"`
	MaxRetries          int     `mapstructure:"max_retries"`
	EnableDLQ           *bool   `mapstructure:"enable_dlq"`

	Compression  string   `mapstructure:"compression"`
	RequiredAcks int      `mapstructure:"required_acks"`
	Brokers      []string `mapstructure:"brokers"`

	Host          string `mapstructure:"host"`
	Port          int    `mapstructure:"port"`
	VHost         string `mapstructure:"vhost"`
	Username      string `mapstructure:"username"`
	Password      string `mapstructure:"password"`
	MessageTTL    int    `mapstructure:"message_ttl"`
	DeliveryMode  int    `mapstructure:"delivery_mode"`
	QueueType     string `mapstructure:"queue_type"`
	PrefetchCount int    `mapstructure:"prefetch_count"`

	ProjectID    string `mapstructure:"project_id"`
	EmulatorHost string `mapstructure:"emulator_host"`
}

// RawTopicConfig is the shape of one entry under message_publishers /
// message_consumers (besides the reserved "general_mq_config" key).
type RawTopicConfig struct {
	Name         string        `mapstructure:"name"`
	Type         BrokerType    `mapstructure:"type"`
	Subscription string        `mapstructure:"subscription"` // consumers only
	MQConfig     TopicOverride `mapstructure:"mq_config"`
}

// MQPublishersConfig is the raw message_publishers block, before per-topic
// merging.
type MQPublishersConfig struct {
	GeneralMQConfig GeneralMQConfig           `mapstructure:"general_mq_config"`
	Topics          map[string]RawTopicConfig `mapstructure:"-"` // filled by Load via viper.Sub
}

// MQConsumersConfig is the raw message_consumers block.
type MQConsumersConfig struct {
	Subscription    string                    `mapstructure:"subscription"`
	GeneralMQConfig GeneralMQConfig           `mapstructure:"general_mq_config"`
	Topics          map[string]RawTopicConfig `mapstructure:"-"`
}

// ResolvedTopic is one topic's fully-merged configuration: the matching
// broker-specific block from GeneralMQConfig with the topic's own
// mq_config overrides applied on top. internal/repository/mq consumes this
// directly — no further merging needed downstream.
type ResolvedTopic struct {
	Key          string // logical key, e.g. "user_created"
	Name         string // physical topic/queue name on the broker
	Type         BrokerType
	Subscription string // consumer group / queue name (consumers only)

	Kafka    KafkaConfig
	RabbitMQ RabbitMQConfig
	PubSub   PubSubConfig

	ConcurrentConsumers int
	ExponentialBackoff  bool
	InitialInterval     int
	MaxInterval         int
	Multiplier          float64
	MaxRetries          int
	EnableDLQ           bool
}

// Resolve merges general with each topic's override and returns one
// ResolvedTopic per entry, keyed by the topic's logical key.
func (g GeneralMQConfig) resolve(topics map[string]RawTopicConfig, defaultSubscription string) map[string]ResolvedTopic {
	out := make(map[string]ResolvedTopic, len(topics))
	for key, raw := range topics {
		r := ResolvedTopic{
			Key:                 key,
			Name:                raw.Name,
			Type:                raw.Type,
			Subscription:        firstNonEmpty(raw.Subscription, defaultSubscription),
			Kafka:               g.Kafka,
			RabbitMQ:            g.RabbitMQ,
			PubSub:              g.PubSub,
			ConcurrentConsumers: orInt(raw.MQConfig.ConcurrentConsumers, g.ConcurrentConsumers, 1),
			ExponentialBackoff:  orBool(raw.MQConfig.ExponentialBackoff, g.ExponentialBackoff),
			InitialInterval:     orInt(raw.MQConfig.InitialInterval, g.InitialInterval, 2),
			MaxInterval:         orInt(raw.MQConfig.MaxInterval, g.MaxInterval, 30),
			Multiplier:          orFloat(raw.MQConfig.Multiplier, g.Multiplier, 2),
			MaxRetries:          orInt(raw.MQConfig.MaxRetries, g.MaxRetries, 5),
			EnableDLQ:           orBool(raw.MQConfig.EnableDLQ, g.EnableDLQ),
		}

		ov := raw.MQConfig
		switch raw.Type {
		case BrokerKafka:
			if len(ov.Brokers) > 0 {
				r.Kafka.Brokers = ov.Brokers
			}
			if ov.Compression != "" {
				r.Kafka.Compression = ov.Compression
			}
			if ov.RequiredAcks != 0 {
				r.Kafka.RequiredAcks = ov.RequiredAcks
			}
		case BrokerRabbitMQ:
			if ov.Host != "" {
				r.RabbitMQ.Host = ov.Host
			}
			if ov.Port != 0 {
				r.RabbitMQ.Port = ov.Port
			}
			if ov.VHost != "" {
				r.RabbitMQ.VHost = ov.VHost
			}
			if ov.Username != "" {
				r.RabbitMQ.Username = ov.Username
			}
			if ov.Password != "" {
				r.RabbitMQ.Password = ov.Password
			}
			if ov.MessageTTL != 0 {
				r.RabbitMQ.MessageTTL = ov.MessageTTL
			}
			if ov.DeliveryMode != 0 {
				r.RabbitMQ.DeliveryMode = ov.DeliveryMode
			}
			if ov.QueueType != "" {
				r.RabbitMQ.QueueType = ov.QueueType
			}
			if ov.PrefetchCount != 0 {
				r.RabbitMQ.PrefetchCount = ov.PrefetchCount
			}
		case BrokerPubSub:
			if ov.ProjectID != "" {
				r.PubSub.ProjectID = ov.ProjectID
			}
			if ov.EmulatorHost != "" {
				r.PubSub.EmulatorHost = ov.EmulatorHost
			}
		}

		out[key] = r
	}
	return out
}

func firstNonEmpty(values ...string) string {
	for _, v := range values {
		if v != "" {
			return v
		}
	}
	return ""
}

func orInt(v, fallback, defaultVal int) int {
	if v != 0 {
		return v
	}
	if fallback != 0 {
		return fallback
	}
	return defaultVal
}

func orFloat(v, fallback, defaultVal float64) float64 {
	if v != 0 {
		return v
	}
	if fallback != 0 {
		return fallback
	}
	return defaultVal
}

func orBool(v *bool, fallback bool) bool {
	if v != nil {
		return *v
	}
	return fallback
}
