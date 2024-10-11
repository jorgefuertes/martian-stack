package server_test

import (
	"fmt"
	"net/http"

	"git.martianoids.com/martianoids/martian-stack/pkg/server"
)

func testErrorHandlerfunc(c server.Ctx, err error) {
	var e server.HttpError
	e, ok := err.(server.HttpError)
	if !ok {
		e = server.HttpError{Code: http.StatusInternalServerError, Msg: err.Error()}
	}
	e.Msg = fmt.Sprintf("TestErrorHandler: %d %s", e.Code, e.Msg)

	if c.AcceptsJSON() {
		_ = c.WithStatus(e.Code).SendJSON(e)
	} else {
		_ = c.WithStatus(e.Code).SendString(e.Error())
	}
}
