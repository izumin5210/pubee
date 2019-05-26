package pubee

import "context"

type PublisherConfig struct {
	PublishOpts       []PublishOption
	Interceptor       Interceptor
	OnFailPublishFunc func(*Message, error)
}

func (c *PublisherConfig) apply(opts []PublisherOption) {
	for _, f := range opts {
		f.ApplyPublisherOption(c)
	}
}

func (c *PublisherConfig) OnFailPublish(msg *Message, err error) {
	if f := c.OnFailPublishFunc; f != nil {
		f(msg, err)
	}
}

type PublishConfig struct {
	Metadata map[string]string
	Marshal  MarshalFunc
}

func (c *PublishConfig) apply(opts []PublishOption) {
	for _, f := range opts {
		f.ApplyPublishOption(c)
	}
}

type Option interface {
	PublisherOption
	PublishOption
}

type PublisherOption interface {
	ApplyPublisherOption(*PublisherConfig)
}

type PublisherOptionFunc func(*PublisherConfig)

func (o PublisherOptionFunc) ApplyPublisherOption(c *PublisherConfig) { o(c) }

type PublishOption interface {
	ApplyPublishOption(*PublishConfig)
}

type BothOptionFunc func(*PublishConfig)

func (o BothOptionFunc) ApplyPublishOption(c *PublishConfig) { o(c) }

func (o BothOptionFunc) ApplyPublisherOption(c *PublisherConfig) {
	c.PublishOpts = append(c.PublishOpts, o)
}

func WithInterceptors(interceptors ...Interceptor) PublisherOption {
	return PublisherOptionFunc(func(c *PublisherConfig) {
		if f := c.Interceptor; f != nil {
			interceptors = append([]Interceptor{f}, interceptors...)
		}
		n := len(interceptors)
		if n == 1 {
			c.Interceptor = interceptors[0]
			return
		}

		// https://github.com/grpc-ecosystem/go-grpc-middleware/blob/v1.0.0/chain.go#L13-L51
		lastI := n - 1
		c.Interceptor = func(ctx context.Context, msg *Message, handler func(context.Context, *Message)) {
			var (
				chainHandler func(context.Context, *Message)
				curI         int
			)

			chainHandler = func(currentCtx context.Context, currentMsg *Message) {
				if curI == lastI {
					handler(currentCtx, currentMsg)
					return
				}
				curI++
				interceptors[curI](currentCtx, currentMsg, chainHandler)
				curI--
			}

			interceptors[0](ctx, msg, chainHandler)
		}
	})
}

func WithMetadata(kv ...string) Option {
	return BothOptionFunc(func(c *PublishConfig) {
		if c.Metadata == nil {
			c.Metadata = map[string]string{}
		}
		for i := 0; i < len(kv)/2; i++ {
			c.Metadata[kv[2*i]] = kv[2*i+1]
		}
	})
}

func WithMetadataMap(meta map[string]string) Option {
	return BothOptionFunc(func(c *PublishConfig) {
		if c.Metadata == nil {
			c.Metadata = meta
			return
		}
		for k, v := range meta {
			c.Metadata[k] = v
		}
	})
}

func WithJSON() Option {
	return WithMarshalFunc(MarshalJSON)
}

func WithProtobuf() Option {
	return WithMarshalFunc(MarshalProtobuf)
}

func WithMarshalFunc(f MarshalFunc) Option {
	return BothOptionFunc(func(c *PublishConfig) { c.Marshal = f })
}

func WithOnFailPublish(f func(*Message, error)) PublisherOption {
	return PublisherOptionFunc(func(c *PublisherConfig) {
		c.OnFailPublishFunc = f
	})
}
