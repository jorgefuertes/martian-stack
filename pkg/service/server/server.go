package server

import (
	"net/http"
	"time"

	"git.martianoids.com/martianoids/martian-stack/pkg/service/logger"
)

type (
	Handler func(c Ctx) error
)

type Server struct {
	srv          *http.Server
	mux          *http.ServeMux
	handlers     []Handler
	errorHandler ErrorHandler
}

func New(host, port string, timeoutSeconds int, log *logger.Service) *Server {
	t := time.Second * time.Duration(timeoutSeconds)
	mux := http.NewServeMux()

	srv := &http.Server{
		Addr:              host + ":" + port,
		Handler:           mux,
		ReadTimeout:       t,
		ReadHeaderTimeout: t,
		WriteTimeout:      t,
	}

	return &Server{
		srv:          srv,
		mux:          mux,
		errorHandler: defaultErrorHandler,
	}
}

func (s *Server) Start() error {
	return s.srv.ListenAndServe()
}

func (s *Server) Use(mw ...Handler) {
	s.handlers = append(s.handlers, mw...)
}

func (s *Server) Route(method, path string, h Handler) {
	s.mux.HandleFunc(method+" "+path, func(w http.ResponseWriter, r *http.Request) {
		c := newCtx(w, r, s.handlers...)
		// execute all the handlers in a "next" chain
		if err := c.Next(); err != nil {
			s.errorHandler(c, err)
		}
	})
}
