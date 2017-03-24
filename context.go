package minitel

import "context"

type minitelKey int

var key minitelKey

// NewContext returns a new Context that includes the Client
func NewContext(ctx context.Context, cli Client) context.Context {
	return context.WithValue(ctx, key, cli)
}

// FromContext retrieves the Client value stored in ctx, if any.
func FromContext(ctx context.Context) (cli Client, ok bool) {
	cli, ok = ctx.Value(key).(Client)
	return
}
