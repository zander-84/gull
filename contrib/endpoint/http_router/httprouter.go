package http_router

import (
	"github.com/julienschmidt/httprouter"
	"github.com/zander-84/gull/endpoint"
	"github.com/zander-84/gull/transport/http"
	http2 "net/http"
)

var _ http2.Handler = (*Router)(nil)

type Router struct {
	engine *httprouter.Router
}

func NewRouter(engine *httprouter.Router) *Router {
	router := new(Router)
	router.engine = engine
	return router
}

func (r *Router) ServeHTTP(res http2.ResponseWriter, req *http2.Request) {
	r.engine.ServeHTTP(res, req)
}
func (r *Router) Endpoint(protocol endpoint.Protocol, method endpoint.Method, path string, e endpoint.HandlerFunc) {
	switch method {

	case endpoint.MethodGet:
		r.engine.GET(path, func(writer http2.ResponseWriter, request *http2.Request, params httprouter.Params) {
			request = setRequest(request, protocol)
			_, _ = e(http.NewHttpContext(writer, request), nil)
		})
	case endpoint.MethodHead:
		r.engine.HEAD(path, func(writer http2.ResponseWriter, request *http2.Request, params httprouter.Params) {
			request = setRequest(request, protocol)
			_, _ = e(http.NewHttpContext(writer, request), nil)
		})
	case endpoint.MethodPost:
		r.engine.POST(path, func(writer http2.ResponseWriter, request *http2.Request, params httprouter.Params) {
			request = setRequest(request, protocol)
			_, _ = e(http.NewHttpContext(writer, request), nil)
		})
	case endpoint.MethodPut:
		r.engine.PUT(path, func(writer http2.ResponseWriter, request *http2.Request, params httprouter.Params) {
			request = setRequest(request, protocol)
			_, _ = e(http.NewHttpContext(writer, request), nil)
		})
	case endpoint.MethodPatch:
		r.engine.PATCH(path, func(writer http2.ResponseWriter, request *http2.Request, params httprouter.Params) {
			request = setRequest(request, protocol)
			_, _ = e(http.NewHttpContext(writer, request), nil)
		})
	case endpoint.MethodDelete:
		r.engine.DELETE(path, func(writer http2.ResponseWriter, request *http2.Request, params httprouter.Params) {
			request = setRequest(request, protocol)
			_, _ = e(http.NewHttpContext(writer, request), nil)
		})
	case endpoint.MethodOptions:
		r.engine.OPTIONS(path, func(writer http2.ResponseWriter, request *http2.Request, params httprouter.Params) {
			request = setRequest(request, protocol)
			_, _ = e(http.NewHttpContext(writer, request), nil)
		})
	default:
	}
}

func setRequest(request *http2.Request, protocol endpoint.Protocol) *http2.Request {
	endpointCtxVal := endpoint.NewCtxVal()
	endpointCtxVal.SetProtocol(protocol)
	return request.WithContext(endpoint.WithContext(request.Context(), endpointCtxVal))
}
