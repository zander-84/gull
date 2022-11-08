package think

import (
	"errors"
	"fmt"
	"google.golang.org/grpc/status"
)

type Error struct {
	Response
	cause error
}

// Unwrap provides compatibility for Go 1.13 error chains.
func (e *Error) Unwrap() error { return e.cause }

func (e *Error) Error() string {
	return fmt.Sprintf("error: code = %d data = %s message = %s metadata = %v cause = %v", e.Code, e.Data, e.Message, e.Metadata, e.cause)
}

// Is matches each error in the chain with the target value.
func (e *Error) Is(err error) bool {
	if se := new(Error); errors.As(err, &se) {
		return se.Code == e.Code && se.Message == e.Message
	}
	return false
}

// WithCause with the underlying cause of the error.
func (e *Error) WithCause(cause error) *Error {
	err := Clone(e)
	err.cause = cause
	return err
}

// WithMetadata with an MD formed by the mapping of key, value.
func (e *Error) WithMetadata(md map[string]string) *Error {
	err := Clone(e)
	err.Metadata = md
	return err
}

// New returns an error object for the code, message.
func New(code Code, bizCode string, message, reason string) *Error {
	return &Error{
		Response: Response{
			Code:    code,
			BizCode: bizCode,
			Message: message,
			Data:    reason,
		},
	}
}

// Clone deep clone error to a new error.
func Clone(err *Error) *Error {
	if err == nil {
		return nil
	}
	metadata := make(map[string]string, len(err.Metadata))
	for k, v := range err.Metadata {
		metadata[k] = v
	}
	return &Error{
		cause: err.cause,
		Response: Response{
			Code:     err.Code,
			Data:     err.Data,
			Message:  err.Message,
			Metadata: metadata,
		},
	}
}

// GetCode returns the http code for an error.
// It supports wrapped errors.
func GetCode(err error) Code {
	if err == nil {
		return CodeSuccess
	}
	return FromError(err).Code
}

// FromError try to convert an error to *Error.
// It supports wrapped errors.
func FromError(err error) *Error {
	if err == nil {
		return nil
	}
	if se := new(Error); errors.As(err, &se) {
		return se
	}
	gs, ok := status.FromError(err)
	if !ok {
		return New(CodeUndefined, "", CodeUndefined.ToString(), err.Error())
	}

	if uint32(gs.Code()) < uint32(MinCode) {
		return New(CodeSystemSpaceError, "", CodeSystemSpaceError.ToString(), gs.Message())
	} else {
		c := Code(gs.Code())
		if c == CodeBizError {
			return New(CodeSystemSpaceError, "", CodeSystemSpaceError.ToString(), gs.Message()) // todo
		} else {
			return New(c, "", c.ToString(), gs.Message())
		}
	}

}
