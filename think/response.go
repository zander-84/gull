package think

type Response struct {
	Code     Code
	BizCode  string
	Message  string
	Metadata map[string]string
	Data     interface{}
}

func NewResponse(code Code, bizCode string, message string, metadata map[string]string, Data interface{}) *Response {
	return &Response{
		Code:     code,
		BizCode:  bizCode,
		Message:  message,
		Metadata: metadata,
		Data:     Data,
	}
}
