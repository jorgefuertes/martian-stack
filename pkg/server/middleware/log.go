package middleware

import (
	"net/http"

	"git.martianoids.com/martianoids/martian-stack/pkg/server"
)

type loggerService interface {
	Request(id, method, path, ip, sessID string, status int, err error)
}

func NewLog(l loggerService) server.Handler {
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

		var sessID string
		if c.Session() != nil {
			sessID = c.Session().ID
		}
		l.Request(c.ID(), c.Method(), c.Path(), c.UserIP(), sessID, code, err)

		return err
	}
}
