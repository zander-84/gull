package http_gin

import (
	"github.com/gin-gonic/gin"
	"github.com/zander-84/gull/endpoint"
	"github.com/zander-84/gull/transport/http"
	http2 "net/http"
)

var _ http2.Handler = (*Router)(nil)

type Router struct {
	ginEngine *gin.Engine
}

func NewRouter(ginEngine *gin.Engine) *Router {
	router := new(Router)
	router.ginEngine = ginEngine

	return router
}

func (r *Router) ServeHTTP(res http2.ResponseWriter, req *http2.Request) {
	r.ginEngine.Handler().ServeHTTP(res, req)
}
func (r *Router) Endpoint(protocol endpoint.Protocol, method endpoint.Method, path string, e endpoint.HandlerFunc) {

	switch method {
	case endpoint.MethodGet:
		r.ginEngine.GET(path, func(ctx *gin.Context) {
			initCtx(ctx, protocol)
			_, _ = e(http.NewHttpContext(ctx.Writer, ctx.Request), nil)
		})
	case endpoint.MethodHead:
		r.ginEngine.OPTIONS(path, func(ctx *gin.Context) {
			initCtx(ctx, protocol)
			_, _ = e(http.NewHttpContext(ctx.Writer, ctx.Request), nil)
		})
	case endpoint.MethodPost:
		r.ginEngine.POST(path, func(ctx *gin.Context) {
			initCtx(ctx, protocol)
			_, _ = e(http.NewHttpContext(ctx.Writer, ctx.Request), nil)
		})
	case endpoint.MethodPut:
		r.ginEngine.PUT(path, func(ctx *gin.Context) {
			initCtx(ctx, protocol)
			_, _ = e(http.NewHttpContext(ctx.Writer, ctx.Request), nil)
		})
	case endpoint.MethodPatch:
		r.ginEngine.PATCH(path, func(ctx *gin.Context) {
			initCtx(ctx, protocol)
			_, _ = e(http.NewHttpContext(ctx.Writer, ctx.Request), nil)
		})
	case endpoint.MethodDelete:
		r.ginEngine.DELETE(path, func(ctx *gin.Context) {
			initCtx(ctx, protocol)
			_, _ = e(http.NewHttpContext(ctx.Writer, ctx.Request), nil)
		})
	case endpoint.MethodOptions:
		r.ginEngine.OPTIONS(path, func(ctx *gin.Context) {
			initCtx(ctx, protocol)
			_, _ = e(http.NewHttpContext(ctx.Writer, ctx.Request), nil)
		})
	default:
		r.ginEngine.Any(path, func(ctx *gin.Context) {
			initCtx(ctx, protocol)
			_, _ = e(http.NewHttpContext(ctx.Writer, ctx.Request), nil)
		})
	}
}

func initCtx(ctx *gin.Context, protocol endpoint.Protocol) {
	endpointCtxVal := endpoint.NewCtxVal()
	endpointCtxVal.SetProtocol(protocol)
	ctx.Request = ctx.Request.WithContext(endpoint.WithContext(ctx.Request.Context(), endpointCtxVal))
}
