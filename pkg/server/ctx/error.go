package ctx

import (
	"net/http"

	"git.martianoids.com/martianoids/martian-stack/pkg/server/servererror"
)

// helper to compose an HttpError to be used as error return
func (c Ctx) Error(code int, message any) servererror.Error {
	var msg string

	switch m := message.(type) {
	case string:
		msg = m
	case error:
		msg = m.Error()
	default:
		msg = http.StatusText(code)
	}

	return servererror.Error{Code: code, Msg: msg}
}
