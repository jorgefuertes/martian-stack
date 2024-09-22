package server

import (
	"context"
	"net/http"
	"time"

	"git.martianoids.com/martianoids/martian-stack/pkg/service/server/web"
)

type Server struct {
	srv          *http.Server
	mux          *http.ServeMux
	handlers     []Handler
	errorHandler ErrorHandler
}

const closeTimeoutSeconds = 30

func New(host, port string, timeoutSeconds int) *Server {
	t := time.Second * time.Duration(timeoutSeconds)
	mux := http.NewServeMux()

	httpSrv := &http.Server{
		Addr:              host + ":" + port,
		Handler:           mux,
		ReadTimeout:       t,
		ReadHeaderTimeout: t,
		WriteTimeout:      t,
	}

	s := &Server{
		srv:          httpSrv,
		mux:          mux,
		handlers:     []Handler{},
		errorHandler: defaultErrorHandler,
	}

	s.Route(web.MethodAny, "/", func(c Ctx) error {
		return c.Error(http.StatusNotFound, ErrNotFound)
	})

	return s
}

func (s *Server) Start() error {
	return s.srv.ListenAndServe()
}

func (s *Server) Stop() error {
	ctx, cancel := context.WithTimeout(context.Background(), closeTimeoutSeconds*time.Second)
	defer cancel()
	return s.srv.Shutdown(ctx)
}

func (s *Server) Use(mw ...Handler) {
	s.handlers = append(s.handlers, mw...)
}

func (s *Server) Route(method web.Method, path string, h Handler) {
	if !web.IsValidMethod(method) {
		method = web.MethodGet
	}

	if !web.IsMethodAny(method) {
		path = method.String() + " " + path
	}

	s.mux.HandleFunc(path, func(w http.ResponseWriter, r *http.Request) {
		mw := s.handlers

		if isRootPath(path) {
			mw = append(mw, notFoundMiddleware)
		}

		c := NewCtx(w, r, append(mw, h)...)

		// execute all the handlers in a "next" chain
		if err := c.Next(); err != nil {
			s.errorHandler(c, err)
		}
	})
}

func (s *Server) ErrorHandler(h ErrorHandler) {
	s.errorHandler = h
}
