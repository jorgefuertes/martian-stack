package server

import (
	"net/http"

	"git.martianoids.com/martianoids/martian-stack/pkg/helper"
	"git.martianoids.com/martianoids/martian-stack/pkg/server/ctx"
	"git.martianoids.com/martianoids/martian-stack/pkg/server/web"
)

// method: httpconst.Method
// path: path to be handled, params can be defined as :param or {param}
func (s *Server) Route(method web.Method, path string, h ctx.Handler) {
	if !web.IsValidMethod(method) {
		method = web.MethodGet
	}

	if !web.IsMethodAny(method) {
		path = method.String() + " " + path
	}

	// replace :param with {param}
	path = helper.ReplacePathParams(path)

	s.mux.HandleFunc(path, func(w http.ResponseWriter, r *http.Request) {
		mw := s.handlers

		if helper.IsRootPath(path) {
			mw = append(mw, notFoundMiddleware)
		}

		c := ctx.New(w, r, append(mw, h)...)

		// execute all the handlers in a "next" chain
		if err := c.Next(); err != nil {
			s.errorHandler(c, err)
		}
	})
}
