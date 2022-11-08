package think

func ErrUndefined(reason string) error {
	return New(CodeUndefined, "", CodeUndefined.ToString(), reason)
}

func IsErrUndefined(err error) bool {
	return GetCode(err) == CodeUndefined
}
func ErrSystemSpace(reason string) error {
	return New(CodeSystemSpaceError, "", CodeSystemSpaceError.ToString(), reason)
}

func IsErrSystemSpace(err error) bool {
	return GetCode(err) == CodeSystemSpaceError
}
