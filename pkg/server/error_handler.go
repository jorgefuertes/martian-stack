package server

import (
	"github.com/jorgefuertes/martian-stack/pkg/server/ctx"
	"github.com/jorgefuertes/martian-stack/pkg/server/servererror"
	"github.com/jorgefuertes/martian-stack/pkg/server/view"
)

type ErrorHandler func(c ctx.Ctx, err error)

func defaultErrorHandler(c ctx.Ctx, err error) {
	e, ok := err.(servererror.Error)
	if !ok {
		e = servererror.New().WithMsg(err.Error())
	}

	if c.AcceptsJSON() {
		_ = c.WithStatus(e.Code).SendJSON(e)
	} else if c.AcceptsPlainText() {
		_ = c.WithStatus(e.Code).SendString(e.Error())
	} else {
		_ = c.WithStatus(e.Code).Render(view.Error(e))
	}
}

// returns a 404 error if the request path is different from "/"
// it should be used with a "/" route because that route acts as a catch-all,
// and overwrites the server previous cath-all.
func notFoundMiddleware(c ctx.Ctx) error {
	if c.Path() != "/" {
		return servererror.ErrNotFound
	}

	return c.Next()
}
