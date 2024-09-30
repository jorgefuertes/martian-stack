package server

import (
	"context"
	"net/http"
	"time"

	"git.martianoids.com/martianoids/martian-stack/pkg/httpconst"
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

	s.Route(httpconst.MethodAny, "/", func(c Ctx) error {
		return c.Error(http.StatusNotFound, ErrNotFound)
	})

	s.Route(httpconst.MethodGet, "/server/ready", func(c Ctx) error {
		return c.SendString("OK")
	})

	return s
}

func (s *Server) Start() error {
	return s.srv.ListenAndServe()
}

func (s *Server) IsReady() bool {
	req, err := http.NewRequest("GET", "http://"+s.srv.Addr+"/server/ready", nil)
	if err != nil {
		return false
	}

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return false
	}

	return resp.StatusCode == http.StatusOK
}

func (s *Server) WaitUntilReady() {
	t := time.Now()
	for !s.IsReady() {
		if time.Since(t) > time.Second*10 {
			return
		}
		time.Sleep(time.Second)
	}
}

func (s *Server) Stop() error {
	ctx, cancel := context.WithTimeout(context.Background(), closeTimeoutSeconds*time.Second)
	defer cancel()
	return s.srv.Shutdown(ctx)
}

func (s *Server) Use(mw ...Handler) {
	s.handlers = append(s.handlers, mw...)
}

// method: httpconst.Method
// path: path to be handled, params can be defined as :param or {param}
func (s *Server) Route(method httpconst.Method, path string, h Handler) {
	if !httpconst.IsValidMethod(method) {
		method = httpconst.MethodGet
	}

	if !httpconst.IsMethodAny(method) {
		path = method.String() + " " + path
	}

	// replace :param with {param}
	path = replacePathParams(path)

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
