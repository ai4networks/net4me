package node

type ErrorNotImplemented struct {
	Err error
}

func (e *ErrorNotImplemented) Error() string {
	return "not implemented"
}
