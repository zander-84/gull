package grpc

import (
	"context"
	"time"
)

var _ Context = (*wrapper)(nil)

// Context is an HTTP Context.
type Context interface {
	context.Context
}

func NewGrpcContext(ctx context.Context) Context {
	w := &wrapper{ctx: ctx}
	return w
}

type wrapper struct {
	ctx context.Context
}

func (c *wrapper) Deadline() (time.Time, bool) {
	return c.ctx.Deadline()
}

func (c *wrapper) Done() <-chan struct{} {
	return c.ctx.Done()
}

func (c *wrapper) Err() error {
	return c.ctx.Err()
}

func (c *wrapper) Value(key interface{}) interface{} {
	return c.ctx.Value(key)
}
