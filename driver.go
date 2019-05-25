package pubee

import "context"

type Driver interface {
	Publish(context.Context, *Message)
	Close(context.Context) error
	SetCallback(DriverCallback)
}

type DriverCallback interface {
	OnFailPublish(*Message, error)
}
