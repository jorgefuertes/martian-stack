package middleware

import (
	"git.martianoids.com/martianoids/martian-stack/pkg/service/server"
)

type loggerService interface {
	Request(method, path, ip string, status int, err error)
}

func NewLogMiddleware(l loggerService) server.Handler {
	return func(c server.Ctx) error {
		err := c.Next()
		l.Request(c.Method(), c.Path(), c.UserIP(), c.Status(), err)

		return err
	}
}
