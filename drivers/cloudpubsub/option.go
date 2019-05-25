package cloudpubsub

import (
	"cloud.google.com/go/pubsub"
	"google.golang.org/api/option"
)

// Config represents publisher configuration.
type Config struct {
	ClientOpts          []option.ClientOption
	PublishSettingsFunc func(*pubsub.PublishSettings)
	TopicConfig         *pubsub.TopicConfigToUpdate
	CreateTopic         bool
	DeleteTopic         bool
}

func (c *Config) apply(opts []Option) {
	for _, f := range opts {
		f(c)
	}
}

// Option is publisher Option
type Option func(*Config)

// WithClientOptions returns an Option that set option.ClientOption implementation(s).
func WithClientOptions(opts ...option.ClientOption) Option {
	return func(c *Config) {
		c.ClientOpts = append(c.ClientOpts, opts...)
	}
}

// WithPublishSettings returns an Option that set pubsub.PublishSettings to the pubsub.Topic.
func WithPublishSettings(f func(*pubsub.PublishSettings)) Option {
	return func(c *Config) {
		c.PublishSettingsFunc = f
	}
}

// WithUpdateTopicConfig returns an Option that update configuration for pubsub.Topic.
func WithUpdateTopicConfig(cfg pubsub.TopicConfigToUpdate) Option {
	return func(c *Config) {
		c.TopicConfig = &cfg
	}
}

func WithCreateTopicIfNeeded() Option {
	return func(c *Config) {
		c.CreateTopic = true
	}
}

func WithDeleteTopicOnClose() Option {
	return func(c *Config) {
		c.DeleteTopic = true
	}
}
