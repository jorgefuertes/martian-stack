# Martian Stack Framework

This is a framework for building web applications in Go.

## Features

- Dependency injection
- Middleware
- Sessions
- Caching
- Validation
- Error handling
- MongoDB integration
- Redis integration

## Installation

```bash
go get -u git.martianoids.com/martianoids/martian-stack
```

## Usage

### Creating a new project

```bash
mkdir my-project
cd my-project
go mod init my-project
```

### Creating a new server

```go
package main

import (
 "os"

 "git.martianoids.com/martianoids/martian-stack/pkg/server"
 "git.martianoids.com/martianoids/martian-stack/pkg/server/httpconst"
 "git.martianoids.com/martianoids/martian-stack/pkg/server/middleware"
 "git.martianoids.com/martianoids/martian-stack/pkg/service/logger"
)

func main() {
 l := logger.New(os.Stdout, logger.TextFormat, logger.LevelDebug)
 srv := server.New("localhost", "8080", 10)
 logMw := middleware.NewLog(l)
 srv.Use(middleware.NewCors(middleware.NewCorsOptions()), logMw)

 // routes
 registerRoutes(srv)

 // start
 l.From("main", "server").With("host", "localhost", "port", "8080", "timeout", "10").Info("starting server")
 err := srv.Start()
 if err != nil {
  l.From("main", "server").Error(err.Error())
 }
}

func registerRoutes(srv *server.Server) {
 srv.Route(httpconst.MethodGet, "/", func(c server.Ctx) error {
  return c.SendString("Welcome to the Home Page")
 })
}
```

## License

Private so far, but perhaps will be open-sourced soon.
