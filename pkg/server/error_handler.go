package server

import (
	"git.martianoids.com/martianoids/martian-stack/pkg/server/ctx"
	"git.martianoids.com/martianoids/martian-stack/pkg/server/server_error"
	"git.martianoids.com/martianoids/martian-stack/pkg/server/view"
)

type ErrorHandler func(c ctx.Ctx, err error)

func defaultErrorHandler(c ctx.Ctx, err error) {
	e := server_error.New()
	e, ok := err.(server_error.Error)
	if !ok {
		e = server_error.New().WithMsg(err.Error())
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
		return server_error.ErrNotFound
	}

	return c.Next()
}
