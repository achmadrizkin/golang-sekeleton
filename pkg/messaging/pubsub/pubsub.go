// Package pubsub implements pkg/messaging.Publisher and Consumer on top of
// Google Cloud Pub/Sub. Set Config.EmulatorHost (e.g. "localhost:8085") to
// target the local emulator used by docker-compose in development instead
// of real GCP.
package pubsub

import (
	"context"
	"fmt"
	"os"

	"cloud.google.com/go/pubsub"

	"github.com/fauzie/golang-sekeleton/pkg/messaging"
)

// Config describes how to reach Pub/Sub (or its emulator) and which topic
// to use.
type Config struct {
	ProjectID    string
	Topic        string
	EmulatorHost string // non-empty targets the local emulator
}

func newClient(ctx context.Context, projectID, emulatorHost string) (*pubsub.Client, error) {
	if emulatorHost != "" {
		if err := os.Setenv("PUBSUB_EMULATOR_HOST", emulatorHost); err != nil {
			return nil, fmt.Errorf("pubsub: set emulator host: %w", err)
		}
	}
	client, err := pubsub.NewClient(ctx, projectID)
	if err != nil {
		return nil, fmt.Errorf("pubsub: new client: %w", err)
	}
	return client, nil
}

// ensureTopic returns the named topic, creating it if it does not exist yet
// (convenient for local/dev environments; in production the topic is
// expected to already exist and this is a cheap existence check).
func ensureTopic(ctx context.Context, client *pubsub.Client, name string) (*pubsub.Topic, error) {
	topic := client.Topic(name)
	ok, err := topic.Exists(ctx)
	if err != nil {
		return nil, fmt.Errorf("pubsub: check topic %q: %w", name, err)
	}
	if !ok {
		if topic, err = client.CreateTopic(ctx, name); err != nil {
			return nil, fmt.Errorf("pubsub: create topic %q: %w", name, err)
		}
	}
	return topic, nil
}

type publisher struct {
	client *pubsub.Client
	topic  *pubsub.Topic
}

// NewPublisher connects and ensures cfg.Topic exists.
func NewPublisher(ctx context.Context, cfg Config) (messaging.Publisher, error) {
	client, err := newClient(ctx, cfg.ProjectID, cfg.EmulatorHost)
	if err != nil {
		return nil, err
	}
	topic, err := ensureTopic(ctx, client, cfg.Topic)
	if err != nil {
		_ = client.Close()
		return nil, err
	}
	return &publisher{client: client, topic: topic}, nil
}

func (p *publisher) Publish(ctx context.Context, topic string, value []byte) error {
	return p.PublishWithMetadata(ctx, topic, value, nil)
}

func (p *publisher) PublishWithMetadata(ctx context.Context, topic string, value []byte, headers map[string]string) error {
	t := p.topic
	if topic != "" && topic != p.topic.ID() {
		var err error
		if t, err = ensureTopic(ctx, p.client, topic); err != nil {
			return err
		}
	}
	result := t.Publish(ctx, &pubsub.Message{Data: value, Attributes: headers})
	if _, err := result.Get(ctx); err != nil {
		return fmt.Errorf("pubsub: publish to %q: %w", topic, err)
	}
	return nil
}

func (p *publisher) Close() error {
	p.topic.Stop()
	return p.client.Close()
}

func (p *publisher) Ping(ctx context.Context) error {
	_, err := p.topic.Exists(ctx)
	if err != nil {
		return fmt.Errorf("pubsub: ping: %w", err)
	}
	return nil
}
