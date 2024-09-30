package main

import (
	"os"

	"git.martianoids.com/martianoids/martian-stack/pkg/middleware"
	"git.martianoids.com/martianoids/martian-stack/pkg/server"
	"git.martianoids.com/martianoids/martian-stack/pkg/service/logger"
)

const (
	host           = "localhost"
	port           = "8080"
	timeoutSeconds = 15
)

func main() {
	l := logger.New(os.Stdout, logger.TextFormat, logger.LevelDebug)
	srv := server.New(host, port, timeoutSeconds)
	logMw := middleware.NewLogMiddleware(l)
	srv.Use(middleware.NewCorsHandler(middleware.NewCorsOptions()), logMw)

	// test routes
	registerRoutes(srv)

	// background start
	l.From("main", "server").With("host", host, "port", port, "timeout", timeoutSeconds).Info("starting server")
	err := srv.Start()
	if err != nil {
		l.From("main", "server").Error(err.Error())
	}
}

func registerRoutes(srv *server.Server) {

}
