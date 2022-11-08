package endpoint

type Method string
type Protocol string

const (
	MethodGet     Method = "GET"
	MethodHead    Method = "HEAD"
	MethodPost    Method = "POST"
	MethodPut     Method = "PUT"
	MethodPatch   Method = "PATCH" // RFC 5789
	MethodDelete  Method = "DELETE"
	MethodOptions Method = "OPTIONS"

	MethodConnect Method = "CONNECT"
	MethodTrace   Method = "TRACE"

	Http   Protocol = "HTTP"
	Grpc   Protocol = "GRPC"
	Custom Protocol = "CUSTOM"
	Empty  Protocol = "EMPTY"
)

func (p Protocol) IsHttp() bool {
	return p == Http
}

func (p Protocol) IsEmpty() bool {
	return p == Empty
}

func (p Protocol) IsGrpc() bool {
	return p == Grpc
}

func (p Protocol) IsCustom() bool {
	return p == Custom
}

func inProtocols(s Protocol, in []Protocol) bool {
	l := len(in)
	if l < 1 {
		return false
	}
	for i := 0; i < l; i++ {
		if in[i] == s {
			return true
		}
	}
	return false

}
