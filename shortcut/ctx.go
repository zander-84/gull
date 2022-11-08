package shortcut

import (
	"context"
	"github.com/zander-84/gull/transport/grpc"
	"github.com/zander-84/gull/transport/http"
)

func IsCtxHttp(ctx context.Context) (http.Context, bool) {
	httpCtx, ok := ctx.(http.Context)
	return httpCtx, ok
}

func IsCtxGrpc(ctx context.Context) (grpc.Context, bool) {
	grpcCtx, ok := ctx.(grpc.Context)
	return grpcCtx, ok
}
