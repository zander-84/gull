package endpoint

import (
	"context"
	"errors"
	"log"
	"strings"
)

// Rmc Resource Management Center
type Rmc interface {
	Group(prefix string, options ...Options) Rmc
	Use(options ...Options) Rmc
	Endpoint(ps []Protocol, method Method, path string, e HandlerFunc, dec DecodeRequestFunc, enc EncodeResponseFunc, options ...Options)
	Proxy(proxy ProxyEndpoint, protocol Protocol)
	GetEndpoint(p Protocol, method Method, path string) (HandlerFunc, error)
	MustGetEndpoint(method Method, path string) HandlerFunc
}

type Conf struct {
	Path       string
	Method     Method
	ps         []Protocol
	Middleware Middleware

	HandlerFunc    HandlerFunc
	RecoverEncoder RecoverEncoder

	DecFunc DecodeRequestFunc
	EncFunc EncodeResponseFunc

	ErrorEncoder ErrorEncoder
}

func newRmcConf() Conf {
	out := Conf{
		DecFunc: func(ctx context.Context, protocol Protocol, request interface{}) (response interface{}, err error) {
			return
		},
	}
	return out
}

type Options func(*Conf)

func OptionsMiddleware(m ...Middleware) Options {
	return func(rmc *Conf) {
		rmc.Middleware = ChainMerge(rmc.Middleware, m...)
	}
}
func OptionsDec(Dec DecodeRequestFunc) Options {
	return func(rmc *Conf) {
		rmc.DecFunc = Dec
	}
}

func OptionsEnc(Enc EncodeResponseFunc) Options {
	return func(rmc *Conf) {
		rmc.EncFunc = Enc
	}
}

func OptionsErrorEncoder(ee ErrorEncoder) Options {
	return func(rmc *Conf) {
		if ee != nil {
			rmc.ErrorEncoder = ee
		}
	}
}

func OptionsRecoverEncoder(re RecoverEncoder) Options {
	return func(rmc *Conf) {
		rmc.RecoverEncoder = re
	}
}

// rmc Resource Management Center
type rmc struct {
	conf      Conf
	endpoints map[string]Conf
}

func NewRmc() Rmc {
	return &rmc{
		conf:      newRmcConf(),
		endpoints: make(map[string]Conf, 0),
	}
}

func (r *rmc) copy() *rmc {
	nr := new(rmc)
	nr.conf = r.conf
	nr.endpoints = r.endpoints
	return nr
}

func (r *rmc) Group(prefix string, options ...Options) Rmc {
	nr := r.copy()
	nr.conf.Path += prefix
	for _, v := range options {
		v(&nr.conf)
	}
	return nr
}

func (r *rmc) Use(options ...Options) Rmc {
	nr := r.copy()
	for _, v := range options {
		v(&nr.conf)
	}
	return nr
}

func (r *rmc) GetEndpoint(p Protocol, method Method, path string) (HandlerFunc, error) {
	conf, err := r.getConfig(method, path)
	if err != nil {
		return nil, err
	}
	if inProtocols(p, conf.ps) {
		return r._endpoint(conf.HandlerFunc, conf.DecFunc, conf.EncFunc, conf.Middleware, conf.ErrorEncoder, conf.RecoverEncoder), nil
	}
	return nil, errors.New("404")
}

func (r *rmc) MustGetEndpoint(method Method, path string) HandlerFunc {
	conf, err := r.getConfig(method, path)
	if err != nil {
		panic("miss endpoint method: 【" + string(method) + "】 path: 【" + path + "】")
	}
	return r._endpoint(conf.HandlerFunc, conf.DecFunc, conf.EncFunc, conf.Middleware, conf.ErrorEncoder, conf.RecoverEncoder)
}

func (r *rmc) Endpoint(ps []Protocol, method Method, path string, hf HandlerFunc, dec DecodeRequestFunc, enc EncodeResponseFunc, options ...Options) {
	key := Key(method, path)
	if _, ok := r.endpoints[key]; ok {
		log.Panicf("路径已经注册 %s", key)
	}

	nr := r.copy()

	for _, v := range options {
		v(&nr.conf)
	}
	nr.conf.DecFunc = dec
	nr.conf.EncFunc = enc
	nr.conf.Method = method
	nr.conf.ps = ps
	nr.conf.Path += path
	nr.conf.HandlerFunc = hf

	if nr.conf.Middleware == nil {
		nr.conf.Middleware = func(handlerFunc HandlerFunc) HandlerFunc {
			return handlerFunc
		}
	}
	nr.endpoints[key] = nr.conf
}

func (r *rmc) getConfig(method Method, path string) (*Conf, error) {
	key := Key(method, path)
	conf, ok := r.endpoints[key]
	if !ok {
		return nil, errors.New("404")
	}
	return &conf, nil

}
func (r *rmc) _endpoint(hf HandlerFunc, dec DecodeRequestFunc, enc EncodeResponseFunc, middleware Middleware, errorEncoder ErrorEncoder, recoverEncoder RecoverEncoder) HandlerFunc {
	return func(ctx context.Context, request interface{}) (interface{}, error) {
		protocol := MustGetCtxVal(ctx).GetProtocol()
		if recoverEncoder != nil {
			defer recoverEncoder(ctx, protocol)
		}

		resp, err := middleware(func(hf HandlerFunc) HandlerFunc {
			return func(ctx context.Context, data interface{}) (interface{}, error) {
				var err error

				if dec != nil {
					if data, err = dec(ctx, protocol, data); err != nil {
						return nil, err
					}
				}

				if data, err = hf(ctx, data); err != nil {
					return nil, err
				}

				return data, nil
			}
		}(hf))(ctx, request)

		if err == nil && enc != nil {
			resp, err = enc(ctx, protocol, resp)
		}

		if err != nil && errorEncoder != nil {
			errorEncoder(ctx, protocol, err)
		}

		return resp, err
	}
}

func (r *rmc) Proxy(proxy ProxyEndpoint, protocol Protocol) {
	for _, v := range r.endpoints {
		conf, err := r.getConfig(v.Method, v.Path)
		if err != nil {
			return
		}
		if inProtocols(protocol, conf.ps) {
			proxy(protocol, v.Method, v.Path, r._endpoint(conf.HandlerFunc, v.DecFunc, v.EncFunc, conf.Middleware, v.ErrorEncoder, conf.RecoverEncoder))
		}

	}
}

func Key(method Method, path string) string {
	return string(method) + ":" + path
}

func parseKey(key string) (Method, string) {
	data := strings.Split(key, ":")
	if len(data) < 2 {
		return Method(data[0]), ""
	}

	return Method(data[0]), strings.Join(data[1:], ":")
}
