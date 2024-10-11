package server_test

import (
	"fmt"
	"net/http"

	"git.martianoids.com/martianoids/martian-stack/pkg/server"
	"git.martianoids.com/martianoids/martian-stack/pkg/server/httpconst"
)

func registerRoutes(srv *server.Server) {
	srv.Route(httpconst.MethodGet, "/", func(c server.Ctx) error {
		return c.SendString("Welcome to the Home Page")
	})

	srv.Route(httpconst.MethodGet, "/hello", func(c server.Ctx) error {
		return c.SendString("Hello, World!")
	})

	srv.Route(httpconst.MethodGet, "/error/500", func(c server.Ctx) error {
		return c.Error(http.StatusInternalServerError, nil)
	})

	srv.Route(httpconst.MethodGet, "/param-test/:name/:age", func(c server.Ctx) error {
		return c.SendString(fmt.Sprintf("Hello, %s! You are %s years old.", c.Param("name"), c.Param("age")))
	})

	srv.Route(httpconst.MethodGet, "/param-query-test", func(c server.Ctx) error {
		return c.SendString(fmt.Sprintf("Hello, %s! You are %s years old.", c.Param("name"), c.Param("age")))
	})

	srv.Route(httpconst.MethodPost, "/post-json-test", func(c server.Ctx) error {
		var u user
		if err := c.UnmarshalBody(&u); err != nil {
			return err
		}

		return c.SendString(fmt.Sprintf("Hello, %s! You are %d years old.", u.Name, u.Age))
	})

	srv.Route(httpconst.MethodGet, "/json-reply-test", func(c server.Ctx) error {
		u := user{Name: "John", Age: 30}

		return c.SendJSON(u)
	})
}
