package server

import (
	"github.com/jorgefuertes/martian-stack/pkg/server/ctx"
	"github.com/jorgefuertes/martian-stack/pkg/server/web"
)

// Group represents a route group with a shared path prefix and middleware.
type Group struct {
	server     *Server
	prefix     string
	middleware []ctx.Handler
}

// Group creates a new route group with the given path prefix and optional middleware.
// The group middleware runs after server-level middleware and before the route handler.
//
//	api := srv.Group("/api/v1", authMiddleware)
//	api.Route(web.MethodGet, "/users", listUsers)    // handles GET /api/v1/users
//	api.Route(web.MethodPost, "/users", createUser)   // handles POST /api/v1/users
func (s *Server) Group(prefix string, middleware ...ctx.Handler) *Group {
	return &Group{
		server:     s,
		prefix:     prefix,
		middleware: middleware,
	}
}

// Route registers a route within this group.
// The final path is prefix + path. The handler chain is:
// server middleware -> group middleware -> handler.
func (g *Group) Route(method web.Method, path string, h ctx.Handler) {
	g.server.route(method, g.prefix+path, g.middleware, h)
}

// Group creates a sub-group with an additional prefix and middleware.
//
//	admin := api.Group("/admin", requireAdmin)
//	admin.Route(web.MethodGet, "/stats", getStats)   // handles GET /api/v1/admin/stats
func (g *Group) Group(prefix string, middleware ...ctx.Handler) *Group {
	combined := make([]ctx.Handler, 0, len(g.middleware)+len(middleware))
	combined = append(combined, g.middleware...)
	combined = append(combined, middleware...)

	return &Group{
		server:     g.server,
		prefix:     g.prefix + prefix,
		middleware: combined,
	}
}
