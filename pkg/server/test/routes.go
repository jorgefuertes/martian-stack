package server_test

import (
	"fmt"
	"net/http"

	"github.com/jorgefuertes/martian-stack/pkg/server"
	"github.com/jorgefuertes/martian-stack/pkg/server/ctx"
	"github.com/jorgefuertes/martian-stack/pkg/server/web"
)

func registerRoutes(srv *server.Server) {
	srv.Route(web.MethodGet, "/", func(c ctx.Ctx) error {
		return c.SendString("Welcome to the Home Page")
	})

	srv.Route(web.MethodGet, "/hello", func(c ctx.Ctx) error {
		return c.SendString("Hello, World!")
	})

	srv.Route(web.MethodGet, "/error/500", func(c ctx.Ctx) error {
		return c.Error(http.StatusInternalServerError, nil)
	})

	srv.Route(web.MethodGet, "/param-test/:name/:age", func(c ctx.Ctx) error {
		return c.SendString(fmt.Sprintf("Hello, %s! You are %s years old.", c.Param("name"), c.Param("age")))
	})

	srv.Route(web.MethodGet, "/param-query-test", func(c ctx.Ctx) error {
		return c.SendString(fmt.Sprintf("Hello, %s! You are %s years old.", c.Param("name"), c.Param("age")))
	})

	srv.Route(web.MethodPost, "/post-json-test", func(c ctx.Ctx) error {
		var u user
		if err := c.UnmarshalBody(&u); err != nil {
			return err
		}

		return c.SendString(fmt.Sprintf("Hello, %s! You are %d years old.", u.Name, u.Age))
	})

	srv.Route(web.MethodGet, "/json-reply-test", func(c ctx.Ctx) error {
		u := user{Name: "John", Age: 30}

		return c.SendJSON(u)
	})
}
