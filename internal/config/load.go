package config

import (
	"fmt"
	"strings"

	"github.com/spf13/viper"
)

const reservedGeneralKey = "general_mq_config"

// Load reads config.yaml (or path, if non-empty) from ./config then ".",
// with AutomaticEnv overrides enabled (SERVER_HTTP_PORT overrides
// server.http_port, etc — see the package doc comment for the list-type
// caveat).
func Load(path string) (*Config, error) {
	v := viper.New()
	v.SetConfigType("yaml")

	if path != "" {
		v.SetConfigFile(path)
	} else {
		v.SetConfigName("config")
		v.AddConfigPath("./config")
		v.AddConfigPath(".")
	}

	v.SetEnvKeyReplacer(strings.NewReplacer(".", "_"))
	v.AutomaticEnv()
	setDefaults(v)

	if err := v.ReadInConfig(); err != nil {
		return nil, fmt.Errorf("config: read config file: %w", err)
	}

	var cfg Config
	if err := v.Unmarshal(&cfg); err != nil {
		return nil, fmt.Errorf("config: unmarshal: %w", err)
	}

	topics, err := loadTopics(v, "message_publishers")
	if err != nil {
		return nil, fmt.Errorf("config: message_publishers: %w", err)
	}
	cfg.MessagePublishers.Topics = topics

	topics, err = loadTopics(v, "message_consumers")
	if err != nil {
		return nil, fmt.Errorf("config: message_consumers: %w", err)
	}
	cfg.MessageConsumers.Topics = topics

	return &cfg, nil
}

// loadTopics reads every key under section except the reserved
// "general_mq_config" key as a RawTopicConfig. Viper's map-of-struct
// unmarshal support is unreliable for deeply nested optional blocks, so
// each topic is unmarshalled individually via v.Sub, which is exact.
func loadTopics(v *viper.Viper, section string) (map[string]RawTopicConfig, error) {
	sub := v.Sub(section)
	if sub == nil {
		return map[string]RawTopicConfig{}, nil
	}

	raw := v.GetStringMap(section)
	topics := make(map[string]RawTopicConfig, len(raw))
	for key := range raw {
		if key == reservedGeneralKey {
			continue
		}
		topicSub := v.Sub(section + "." + key)
		if topicSub == nil {
			continue
		}
		var tc RawTopicConfig
		if err := topicSub.Unmarshal(&tc); err != nil {
			return nil, fmt.Errorf("topic %q: %w", key, err)
		}
		topics[key] = tc
	}
	return topics, nil
}

// ResolvedPublishers merges general_mq_config into every publisher topic.
func (c *Config) ResolvedPublishers() map[string]ResolvedTopic {
	return c.MessagePublishers.GeneralMQConfig.resolve(c.MessagePublishers.Topics, "")
}

// ResolvedConsumers merges general_mq_config into every consumer topic,
// defaulting each topic's Subscription to message_consumers.subscription
// when the topic does not set its own.
func (c *Config) ResolvedConsumers() map[string]ResolvedTopic {
	return c.MessageConsumers.GeneralMQConfig.resolve(c.MessageConsumers.Topics, c.MessageConsumers.Subscription)
}

func setDefaults(v *viper.Viper) {
	v.SetDefault("server.http_port", 4021)
	v.SetDefault("server.grpc_port", 9090)
	v.SetDefault("server.read_timeout_seconds", 30)
	v.SetDefault("server.environment", "development")
	v.SetDefault("database.query_timeout_seconds", 5)
	v.SetDefault("database.max_open_conns", 25)
	v.SetDefault("database.max_idle_conns", 10)
	v.SetDefault("database.conn_max_lifetime_minutes", 5)
	v.SetDefault("logging.level", "info")
	v.SetDefault("logging.encoding", "json")
	v.SetDefault("telemetry.sample_ratio", 1.0)
}
