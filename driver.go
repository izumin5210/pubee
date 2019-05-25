package pubee

import "context"

type Driver interface {
	Publish(context.Context, *Message)
	Close(context.Context) error
}

type DriverCallback interface {
	OnFailPublish(*Message, error)
}

type nopDriverCallback struct{}

func (nopDriverCallback) OnFailPublish(*Message, error) {}

var _ DriverCallback = (*nopDriverCallback)(nil)

type driverCallbackKey struct{}

func SetDriverCallback(ctx context.Context, cb DriverCallback) context.Context {
	return context.WithValue(ctx, driverCallbackKey{}, cb)
}

func GetDriverCallback(ctx context.Context) DriverCallback {
	v := ctx.Value(driverCallbackKey{})
	if v != nil {
		if cb, ok := v.(DriverCallback); ok {
			return cb
		}
	}
	return nopDriverCallback{}
}
