package common

import (
	"context"
	"log/slog"
)

type slogKey struct{}

type ContextHandler struct {
	slog.Handler
}

func (h ContextHandler) Handle(ctx context.Context, r slog.Record) error {
	if attrs, ok := ctx.Value(slogKey{}).([]slog.Attr); ok {
		for _, v := range attrs {
			r.AddAttrs(v)
		}
	}

	return h.Handler.Handle(ctx, r)
}

func AppendCtx(parent context.Context, attr slog.Attr) context.Context {
	if v, ok := parent.Value(slogKey{}).([]slog.Attr); ok {
		v = append(v, attr)
		return context.WithValue(parent, slogKey{}, v)
	}

	v := []slog.Attr{}
	v = append(v, attr)
	return context.WithValue(parent, slogKey{}, v)
}

func NewLoggerWithCtxHandler(parent slog.Handler) *slog.Logger {
	return slog.New(&ContextHandler{Handler: parent})
}
