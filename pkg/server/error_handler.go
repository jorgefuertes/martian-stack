package server

import (
	"net/http"

	"git.martianoids.com/martianoids/martian-stack/pkg/server/view"
)

type ErrorHandler func(c Ctx, err error)

func defaultErrorHandler(c Ctx, err error) {
	var e HttpError
	e, ok := err.(HttpError)
	if !ok {
		e = HttpError{Code: http.StatusInternalServerError, Msg: err.Error()}
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
func notFoundMiddleware(c Ctx) error {
	if c.Path() != "/" {
		return ErrNotFound
	}

	return c.Next()
}
