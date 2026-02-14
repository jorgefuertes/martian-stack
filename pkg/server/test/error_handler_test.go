package server_test

import (
	"fmt"

	"github.com/jorgefuertes/martian-stack/pkg/server/ctx"
	"github.com/jorgefuertes/martian-stack/pkg/server/servererror"
)

func testErrorHandlerfunc(c ctx.Ctx, err error) {
	var e servererror.Error
	e, ok := err.(servererror.Error)
	if !ok {
		e = servererror.New().WithMsg(err.Error())
	}
	e.Msg = fmt.Sprintf("TestErrorHandler: %d %s", e.Code, e.Msg)

	if c.AcceptsJSON() {
		_ = c.WithStatus(e.Code).SendJSON(e)
	} else {
		_ = c.WithStatus(e.Code).SendString(e.Error())
	}
}
