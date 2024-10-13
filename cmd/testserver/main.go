package main

import (
	"os"

	"git.martianoids.com/martianoids/martian-stack/pkg/server"
	"git.martianoids.com/martianoids/martian-stack/pkg/server/ctx"
	"git.martianoids.com/martianoids/martian-stack/pkg/server/middleware"
	"git.martianoids.com/martianoids/martian-stack/pkg/server/web"
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
	logMw := middleware.NewLog(l)
	srv.Use(middleware.NewCors(middleware.NewCorsOptions()), logMw)

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
	srv.Route(web.MethodGet, "/", func(c ctx.Ctx) error {
		return c.SendString("Welcome to the Home Page")
	})
}
