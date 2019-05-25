package pubee

type PublisherConfig struct {
	PublishOpts  []PublishOption
	Interceptors []Interceptor
}

func (c *PublisherConfig) apply(opts []PublisherOption) {
	for _, f := range opts {
		f.ApplyPublisherOption(c)
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
		// TODO: fix order
		c.Interceptors = append(c.Interceptors, interceptors...)
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
