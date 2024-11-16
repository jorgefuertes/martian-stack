package servererror

import (
	"net/http"
	"strconv"
)

type Error struct {
	Code int    `json:"code"`
	Msg  string `json:"msg"`
}

func New() Error {
	return Error{Code: http.StatusInternalServerError, Msg: http.StatusText(http.StatusInternalServerError)}
}

func (e Error) WithCode(code int) Error {
	e.Code = code

	if http.StatusText(code) != "" {
		e.Msg = http.StatusText(code)
	}

	return e
}

func (e Error) WithMsg(msg string) Error {
	e.Msg = msg

	return e
}

func (e Error) Error() string {
	return e.Msg
}

func (e Error) Status() string {
	return strconv.Itoa(e.Code)
}

func (e Error) IsError() bool {
	return e.Code >= 400
}
