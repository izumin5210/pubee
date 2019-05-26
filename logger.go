package pubee

import (
	"context"
	"log"
	"os"
)

// Logger is the interface for logging
type Logger interface {
	Printf(format string, v ...interface{})
	Print(v ...interface{})
}

var defaultErrorLog Logger = log.New(os.Stderr, "[pubee]", log.LstdFlags)

type ctxkeyErrorLog struct{}

func GetErrorLog(ctx context.Context) Logger {
	if v := ctx.Value(ctxkeyErrorLog{}); v != nil {
		if l, ok := v.(Logger); ok {
			return l
		}
	}

	return new(nopLogger)
}

func setErrorLog(ctx context.Context, l Logger) context.Context {
	return context.WithValue(ctx, ctxkeyErrorLog{}, l)
}

type nopLogger struct{}

func (nopLogger) Printf(format string, v ...interface{}) {}
func (nopLogger) Print(v ...interface{})                 {}
