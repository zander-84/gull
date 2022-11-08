package endpoint

import "context"

type Endpoint func(ctx context.Context, request interface{}, dec DecodeRequestFunc, enc EncodeResponseFunc, errorEncoder ErrorEncoder, recoverEncoder RecoverEncoder) (response interface{}, err error)

type HandlerFunc func(ctx context.Context, request interface{}) (response interface{}, err error)

type ProxyEndpoint func(p Protocol, method Method, path string, e HandlerFunc)

type ErrorEncoder func(ctx context.Context, p Protocol, err error)

func WrapError(e map[Protocol]func(ctx context.Context, err error)) ErrorEncoder {
	if e == nil {
		return nil
	}
	return func(ctx context.Context, p Protocol, err error) {
		for k, v := range e {
			if k == p {
				if v == nil {
					break
				}
				v(ctx, err)
				break
			}
		}
	}
}

type RecoverEncoder func(ctx context.Context, p Protocol)

func WrapRecover(rec map[Protocol]func(ctx context.Context)) RecoverEncoder {
	if rec == nil {
		return nil
	}

	return func(ctx context.Context, p Protocol) {
		for k, v := range rec {
			if k == p {
				if v == nil {
					break
				}
				v(ctx)
				break
			}
		}
	}
}

type DecodeRequestFunc func(ctx context.Context, p Protocol, in interface{}) (out interface{}, err error)

type EncodeResponseFunc func(ctx context.Context, p Protocol, in interface{}) (out interface{}, err error)

func WrapDecode(dec map[Protocol]HandlerFunc) DecodeRequestFunc {
	if dec == nil {
		return nil
	}

	return func(ctx context.Context, p Protocol, in interface{}) (out interface{}, err error) {
		for k, v := range dec {
			if k == p {
				if v == nil {
					break
				}
				return v(ctx, in)
			}
		}

		return in, nil
	}
}

func WrapEncode(enc map[Protocol]HandlerFunc) EncodeResponseFunc {
	if enc == nil {
		return nil
	}

	return func(ctx context.Context, p Protocol, in interface{}) (out interface{}, err error) {
		for k, v := range enc {
			if k == p {
				if v == nil {
					break
				}
				return v(ctx, in)
			}
		}

		return in, nil
	}
}
