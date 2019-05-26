package pubee

import (
	"context"
	"sync"

	"github.com/izumin5210/pubee/marshal"
)

type Engine interface {
	Publish(context.Context, interface{}, ...PublishOption) error
	Close(context.Context) error
}

type Message struct {
	Data     []byte
	Metadata map[string]string
	Original interface{}
}

type Interceptor func(context.Context, *Message, func(context.Context, *Message))

func New(d Driver, opts ...Option) Engine {
	cfg := new(Config)
	cfg.apply(opts)
	return &engineImpl{
		driver: d,
		cfg:    cfg,
	}
}

type engineImpl struct {
	driver Driver
	cfg    *Config
	wg     sync.WaitGroup
}

func (p *engineImpl) Publish(ctx context.Context, body interface{}, opts ...PublishOption) error {
	cfg := new(PublishConfig)
	cfg.apply(p.cfg.PublishOpts)
	cfg.apply(opts)

	if cfg.Marshal == nil {
		cfg.Marshal = marshal.Default
	}

	data, err := cfg.Marshal(body)
	if err != nil {
		return err
	}

	var errCh <-chan error
	msg := &Message{Data: data, Metadata: cfg.Metadata, Original: body}

	if f := p.cfg.Interceptor; f == nil {
		errCh = p.driver.Publish(ctx, msg)
	} else {
		f(ctx, msg, func(ctx context.Context, msg *Message) {
			errCh = p.driver.Publish(ctx, msg)
		})
	}

	p.wg.Add(1)
	go func() {
		defer p.wg.Done()
		if err := <-errCh; err != nil {
			if f := p.cfg.OnFailPublishFunc; f != nil {
				f(msg, err)
			}
		}
	}()

	return nil
}

func (p *engineImpl) Close(ctx context.Context) error {
	p.driver.Flush()
	p.wg.Wait()
	return p.driver.Close(ctx)
}