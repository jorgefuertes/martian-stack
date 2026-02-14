package server

import (
	"net/http"

	"github.com/jorgefuertes/martian-stack/pkg/helper"
	"github.com/jorgefuertes/martian-stack/pkg/server/ctx"
	"github.com/jorgefuertes/martian-stack/pkg/server/web"
)

// Route registers a route handler for the given method and path.
// Path params can be defined as :param or {param}.
func (s *Server) Route(method web.Method, path string, h ctx.Handler) {
	s.route(method, path, nil, h)
}

// route is the internal route registration that supports optional extra middleware
// inserted between server-level middleware and the handler.
func (s *Server) route(method web.Method, path string, extra []ctx.Handler, h ctx.Handler) {
	if !web.IsValidMethod(method) {
		method = web.MethodGet
	}

	if !web.IsMethodAny(method) {
		path = method.String() + " " + path
	}

	// replace :param with {param}
	path = helper.ReplacePathParams(path)

	s.mux.HandleFunc(path, func(w http.ResponseWriter, r *http.Request) {
		// build chain: server middleware + [notFound] + extra middleware + handler
		chain := make([]ctx.Handler, 0, len(s.handlers)+len(extra)+2)
		chain = append(chain, s.handlers...)

		if helper.IsRootPath(path) {
			chain = append(chain, notFoundMiddleware)
		}

		chain = append(chain, extra...)
		chain = append(chain, h)

		c := ctx.New(w, r, chain...)

		// propagate request ID to response for tracing
		c.SetHeader(web.HeaderXRequestID, c.ID())

		// execute all the handlers in a "next" chain
		if err := c.Next(); err != nil {
			s.errorHandler(c, err)
		}
	})
}
