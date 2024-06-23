package server

type HttpError struct {
	Code int    `json:"code"`
	Msg  string `json:"msg"`
}

func (e HttpError) Error() string {
	return e.Msg
}

func NewHttpError(code int, err error) HttpError {
	return HttpError{Code: code, Msg: err.Error()}
}

func (e *HttpError) Status() int {
	return e.Code
}

func (e *HttpError) IsError() bool {
	return e.Code >= 400
}
