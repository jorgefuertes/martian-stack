package server

type HttpError struct {
	code int
	msg  string
}

func (e HttpError) Error() string {
	return e.msg
}

func NewHttpError(code int, err error) HttpError {
	return HttpError{code: code, msg: err.Error()}
}

func (e *HttpError) Status() int {
	return e.code
}

func (e *HttpError) IsError() bool {
	return e.code >= 400
}
