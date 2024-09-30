package server

import "strconv"

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

func (e HttpError) Status() string {
	return strconv.Itoa(e.Code)
}

func (e HttpError) IsError() bool {
	return e.Code >= 400
}
