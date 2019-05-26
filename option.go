package pubee

import (
	"context"

	"github.com/izumin5210/pubee/marshal"
)

type Config struct {
	PublishOpts       []PublishOption
	ErrorLog          Logger
	Interceptor       Interceptor
	OnFailPublishFunc func(*Message, error)
}

func (c *Config) apply(opts []Option) {
	for _, f := range opts {
		f.applyOption(c)
	}
}

type Option interface {
	applyOption(*Config)
}

type OptionFunc func(*Config)

func (o OptionFunc) applyOption(c *Config) { o(c) }

type PublishConfig struct {
	Metadata map[string]string
	Marshal  marshal.Func
}

func (c *PublishConfig) apply(opts []PublishOption) {
	for _, f := range opts {
		f.applyPublishOption(c)
	}
}

type PublishOption interface {
	Option
	applyPublishOption(*PublishConfig)
}

type PublishOptionFunc func(*PublishConfig)

func (o PublishOptionFunc) applyOption(c *Config) { c.PublishOpts = append(c.PublishOpts, o) }

func (o PublishOptionFunc) applyPublishOption(c *PublishConfig) { o(c) }

var (
	_ Option        = (OptionFunc)(nil)
	_ Option        = (PublishOptionFunc)(nil)
	_ PublishOption = (PublishOptionFunc)(nil)
)

func WithErrorLog(l Logger) Option {
	return OptionFunc(func(c *Config) {
		c.ErrorLog = l
	})
}

func WithInterceptors(interceptors ...Interceptor) Option {
	return OptionFunc(func(c *Config) {
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

func WithMetadata(kv ...string) PublishOption {
	return PublishOptionFunc(func(c *PublishConfig) {
		if c.Metadata == nil {
			c.Metadata = map[string]string{}
		}
		for i := 0; i < len(kv)/2; i++ {
			c.Metadata[kv[2*i]] = kv[2*i+1]
		}
	})
}

func WithMetadataMap(meta map[string]string) PublishOption {
	return PublishOptionFunc(func(c *PublishConfig) {
		if c.Metadata == nil {
			c.Metadata = meta
			return
		}
		for k, v := range meta {
			c.Metadata[k] = v
		}
	})
}

func WithJSON() PublishOption {
	return WithMarshalFunc(marshal.JSON)
}

func WithProtobuf() PublishOption {
	return WithMarshalFunc(marshal.Protobuf)
}

func WithMarshalFunc(f marshal.Func) PublishOption {
	return PublishOptionFunc(func(c *PublishConfig) { c.Marshal = f })
}

func WithOnFailPublish(f func(*Message, error)) Option {
	return OptionFunc(func(c *Config) {
		c.OnFailPublishFunc = f
	})
}
