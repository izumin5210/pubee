package pubee

import "context"

type Driver interface {
	Publish(context.Context, *Message) <-chan error
	Flush()
	Close(context.Context) error
}
