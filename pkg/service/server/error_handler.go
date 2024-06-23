package server

import "net/http"

type ErrorHandler func(c Ctx, err error)

func defaultErrorHandler(c Ctx, err error) {
	var e HttpError
	e, ok := err.(HttpError)
	if !ok {
		e = HttpError{Code: http.StatusInternalServerError, Msg: err.Error()}
	}

	if c.AcceptsJSON() {
		_ = c.WithStatus(http.StatusInternalServerError).SendJSON(e)
	} else {
		_ = c.WithStatus(http.StatusInternalServerError).SendString(e.Error())
	}
}
