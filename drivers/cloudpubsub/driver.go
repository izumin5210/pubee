package cloudpubsub

import (
	"context"
	"fmt"
	"sync"

	"cloud.google.com/go/pubsub"
	"github.com/izumin5210/pubee"
)

type Driver struct {
	client *pubsub.Client
	topic  *pubsub.Topic
	cfg    *Config
	wg     sync.WaitGroup
}

var _ pubee.Driver = (*Driver)(nil)

func NewDriver(ctx context.Context, projectID, topicID string, opts ...Option) (*Driver, error) {
	cfg := new(Config)
	cfg.apply(opts)

	cli, err := pubsub.NewClient(ctx, projectID, cfg.ClientOpts...)
	if err != nil {
		return nil, err
	}

	topic := cli.Topic(topicID)
	if ok, err := topic.Exists(ctx); err != nil {
		return nil, err
	} else if !ok {
		if !cfg.CreateTopic {
			return nil, fmt.Errorf("%s does not exist", topic.ID())
		}
		topic, err = cli.CreateTopic(ctx, topicID)
		if err != nil {
			return nil, err
		}
	}

	if cfg.TopicConfig != nil {
		_, err := topic.Update(ctx, *cfg.TopicConfig)
		if err != nil {
			return nil, err
		}
	}

	if f := cfg.PublishSettingsFunc; f != nil {
		f(&topic.PublishSettings)
	}

	return &Driver{
		client: cli,
		topic:  topic,
		cfg:    cfg,
	}, nil
}

func (d *Driver) Publish(ctx context.Context, msg *pubee.Message) {
	psMsg := &pubsub.Message{
		Data:       msg.Data,
		Attributes: msg.Metadata,
	}
	res := d.topic.Publish(ctx, psMsg)

	d.wg.Add(1)
	go func() {
		defer d.wg.Done()
		_, err := res.Get(context.Background())
		if err != nil {
			pubee.GetDriverCallback(ctx).OnFailPublish(msg, err)
		}
	}()
}

func (d *Driver) Close(ctx context.Context) error {
	d.topic.Stop()
	d.wg.Wait()
	if d.cfg.DeleteTopic {
		err := d.topic.Delete(ctx)
		if err != nil {
			// TODO
		}
	}
	err := d.client.Close()
	return err
}
