package server

import (
	"context"
	"crypto/tls"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/jorgefuertes/martian-stack/pkg/server/ctx"
	"github.com/jorgefuertes/martian-stack/pkg/server/servererror"
	"github.com/jorgefuertes/martian-stack/pkg/server/web"
)

type Server struct {
	srv          *http.Server
	mux          *http.ServeMux
	handlers     []ctx.Handler
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
		handlers:     []ctx.Handler{},
		errorHandler: defaultErrorHandler,
	}

	s.Route(web.MethodAny, "/", func(c ctx.Ctx) error {
		return c.Error(http.StatusNotFound, servererror.ErrNotFound)
	})

	s.Route(web.MethodGet, "/server/ready", func(c ctx.Ctx) error {
		return c.SendString("OK")
	})

	return s
}

func (s *Server) Start() error {
	return s.srv.ListenAndServe()
}

// StartTLS starts the server with TLS using the provided certificate and key files.
func (s *Server) StartTLS(certFile, keyFile string) error {
	return s.srv.ListenAndServeTLS(certFile, keyFile)
}

// SetTLSConfig sets a custom TLS configuration on the server.
// Call this before StartTLS or ListenAndShutdownTLS for advanced
// scenarios like mutual TLS or custom cipher suites.
func (s *Server) SetTLSConfig(cfg *tls.Config) {
	s.srv.TLSConfig = cfg
}

// ListenAndShutdown starts the server and blocks until a SIGINT or SIGTERM
// signal is received, then performs a graceful shutdown. The optional onShutdown
// callbacks are invoked after the HTTP server stops (use them to close databases,
// flush logs, etc.). Returns nil when the server shuts down cleanly.
func (s *Server) ListenAndShutdown(onShutdown ...func()) error {
	errCh := make(chan error, 1)

	go func() {
		if err := s.srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			errCh <- err
		}
	}()

	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)

	select {
	case err := <-errCh:
		return err
	case <-quit:
	}

	shutdownErr := s.Stop()

	for _, fn := range onShutdown {
		fn()
	}

	return shutdownErr
}

// ListenAndShutdownTLS is like ListenAndShutdown but starts the server with TLS.
func (s *Server) ListenAndShutdownTLS(certFile, keyFile string, onShutdown ...func()) error {
	errCh := make(chan error, 1)

	go func() {
		if err := s.srv.ListenAndServeTLS(certFile, keyFile); err != nil && err != http.ErrServerClosed {
			errCh <- err
		}
	}()

	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)

	select {
	case err := <-errCh:
		return err
	case <-quit:
	}

	shutdownErr := s.Stop()

	for _, fn := range onShutdown {
		fn()
	}

	return shutdownErr
}

func (s *Server) IsReady() bool {
	req, err := http.NewRequest("GET", "http://"+s.srv.Addr+"/server/ready", nil)
	if err != nil {
		return false
	}

	client := &http.Client{Timeout: 2 * time.Second}

	resp, err := client.Do(req)
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

func (s *Server) Use(mw ...ctx.Handler) {
	s.handlers = append(s.handlers, mw...)
}

func (s *Server) ErrorHandler(h ErrorHandler) {
	s.errorHandler = h
}
