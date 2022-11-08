package endpoint

import (
	"context"
	"github.com/zander-84/gull/tool"
)

type CtxVal struct {
	data     *tool.ConcurrentMap
	protocol Protocol
}

type endpointKey struct{}

func MustGetCtxVal(ctx context.Context) *CtxVal {
	v, ok := ctx.Value(endpointKey{}).(*CtxVal)
	if !ok {
		panic("err ctx")
	}
	return v
}

func NewCtxVal() *CtxVal {
	ctx := new(CtxVal)
	ctx.data = tool.NewConcurrentMap()
	ctx.protocol = Empty
	return ctx
}

func WithContext(ctx context.Context, endpointCtx *CtxVal) context.Context {
	if endpointCtx == nil {
		endpointCtx = NewCtxVal()
	}
	return context.WithValue(ctx, endpointKey{}, endpointCtx)
}

func (ctx *CtxVal) SetProtocol(protocol Protocol) {
	ctx.protocol = protocol
}
func (ctx *CtxVal) GetProtocol() Protocol {
	return ctx.protocol
}
