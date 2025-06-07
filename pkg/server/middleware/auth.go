package middleware

import (
	"net/http"
	"strings"

	"git.martianoids.com/martianoids/martian-stack/pkg/server/adapter"
	"git.martianoids.com/martianoids/martian-stack/pkg/server/ctx"
	"git.martianoids.com/martianoids/martian-stack/pkg/server/web"
)

type AccountRepository interface {
	Exists(id string) bool
	Get(id string) (*adapter.Account, error)
	GetByEmail(email string) (*adapter.Account, error)
	GetByUsername(email string) (*adapter.Account, error)
	Create(a *adapter.Account) error
	Update(a *adapter.Account) error
	Delete(id string) error
}

type Rule struct {
	Method     web.Method
	PathPrefix string
	Allowed    []string
}

func NewSessionAuth(r AccountRepository, rules ...Rule) ctx.Handler {
	return func(c ctx.Ctx) error {
		if c.Session() == nil {
			return c.Error(http.StatusUnauthorized, "session not started")
		}

		if !strings.HasPrefix(c.Path(), "/auth") {
			return c.Next()
		}

		if c.Path() == "/auth/login" {
			return c.SendString("login")
		}

		if c.Path() == "/auth/logout" {
			return c.SendString("logout")
		}

		return c.Next()
	}
}
