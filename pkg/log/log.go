package log

import (
	"context"
	"log"

	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

var logger *zap.SugaredLogger

type logKey struct{}

func FromContext(ctx context.Context) *zap.SugaredLogger {
	l := ctx.Value(logKey{})
	if l != nil {
		return l.(*zap.SugaredLogger)
	}
	return logger
}

func WithLogger(ctx context.Context, logger *zap.SugaredLogger) context.Context {
	return context.WithValue(ctx, logKey{}, logger)
}

func UpdateContext(ctx context.Context, fields ...interface{}) (context.Context, *zap.SugaredLogger) {
	l := FromContext(ctx).With(fields...)
	ctx = WithLogger(ctx, l)
	return ctx, l
}

func init() {
	l, err := zap.NewDevelopment(zap.WithCaller(false), zap.AddStacktrace(zapcore.FatalLevel))
	if err != nil {
		log.Fatal(err)
	}
	logger = l.Sugar()
}
