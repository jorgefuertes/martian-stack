package server_test

import (
	"fmt"

	"git.martianoids.com/martianoids/martian-stack/pkg/server/ctx"
	"git.martianoids.com/martianoids/martian-stack/pkg/server/server_error"
)

func testErrorHandlerfunc(c ctx.Ctx, err error) {
	var e server_error.Error
	e, ok := err.(server_error.Error)
	if !ok {
		e = server_error.New().WithMsg(err.Error())
	}
	e.Msg = fmt.Sprintf("TestErrorHandler: %d %s", e.Code, e.Msg)

	if c.AcceptsJSON() {
		_ = c.WithStatus(e.Code).SendJSON(e)
	} else {
		_ = c.WithStatus(e.Code).SendString(e.Error())
	}
}
