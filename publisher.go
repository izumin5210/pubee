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
