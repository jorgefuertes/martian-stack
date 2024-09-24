package middleware

import (
	"net/http"

	"git.martianoids.com/martianoids/martian-stack/pkg/service/server"
)

type loggerService interface {
	Request(method, path, ip string, status int, err error)
}

func NewLogMiddleware(l loggerService) server.Handler {
	return func(c server.Ctx) error {
		err := c.Next()
		code := c.Status()

		if err != nil {
			e, ok := err.(server.HttpError)
			if ok {
				code = e.Code
			} else {
				if code == http.StatusOK {
					code = http.StatusInternalServerError
				}
			}
		}

		l.Request(c.Method(), c.Path(), c.UserIP(), code, err)

		return err
	}
}
