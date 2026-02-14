package middleware

import (
	"net/http"

	"github.com/jorgefuertes/martian-stack/pkg/server/ctx"
	"github.com/jorgefuertes/martian-stack/pkg/server/servererror"
)

type loggerService interface {
	Request(id, method, path, ip, sessID string, status int, err error)
}

func NewLog(l loggerService) ctx.Handler {
	return func(c ctx.Ctx) error {
		err := c.Next()
		code := c.Status()

		if err != nil {
			e, ok := err.(servererror.Error)
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
