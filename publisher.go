package pubee

import (
	"context"
)

type Publisher interface {
	Publish(context.Context, interface{}, ...PublishOption) error
	Close(context.Context) error
}

type Message struct {
	Data     []byte
	Metadata map[string]string
	Original interface{}
}

type Interceptor func(context.Context, Message, func(context.Context, Message))

func NewPublisher(d Driver, opts ...PublisherOption) Publisher {
	cfg := new(PublisherConfig)
	cfg.apply(opts)
	return &publisherImpl{
		driver: d,
		cfg:    cfg,
	}
}

type publisherImpl struct {
	driver Driver
	cfg    *PublisherConfig
}

func (p *publisherImpl) Publish(ctx context.Context, body interface{}, opts ...PublishOption) error {
	cfg := new(PublishConfig)
	cfg.apply(p.cfg.PublishOpts)
	cfg.apply(opts)

	if cfg.Marshal == nil {
		cfg.Marshal = MarshalDefault
	}

	data, err := cfg.Marshal(body)
	if err != nil {
		return err
	}

	ctx = SetDriverCallback(ctx, p.cfg)

	p.driver.Publish(ctx, &Message{Data: data, Metadata: cfg.Metadata, Original: body})

	return nil
}

func (p *publisherImpl) Close(ctx context.Context) error {
	return p.driver.Close(ctx)
}
