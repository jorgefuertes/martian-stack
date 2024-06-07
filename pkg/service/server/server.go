package server

import (
	"fmt"
	"net/http"
)

type Handler func(c Ctx) error
type ErrorHandler func(c Ctx, err error)

type Server struct {
	port           int
	address        string
	timeoutSeconds int
	router         *http.ServeMux
	middlewares    []Handler
	errorHandler   ErrorHandler
}

func New(address string, port, timeoutSeconds int) *Server {
	router := http.NewServeMux()

	return &Server{
		router:         router,
		address:        address,
		port:           port,
		timeoutSeconds: timeoutSeconds,
	}
}

func (s *Server) Start() {
	addr := fmt.Sprintf("%s:%d", s.address, s.port)
	http.ListenAndServe(addr, s.router)
}

func (s *Server) Route(method, path string, h Handler) {
	s.router.HandleFunc(method+" "+path, func(w http.ResponseWriter, r *http.Request) {
		handlers := append(s.middlewares, h)
		c := newCtx(w, r, handlers...)
		// execute all the handlers in a "next" chain
		if err := c.Next(); err != nil {
			s.errorHandler(c, err)
		}
	})
}
