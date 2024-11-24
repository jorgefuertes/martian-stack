package ctx

import (
	"context"
	"net/http"
	"slices"

	"git.martianoids.com/martianoids/martian-stack/pkg/server/adapter"
	"git.martianoids.com/martianoids/martian-stack/pkg/server/session"
	"git.martianoids.com/martianoids/martian-stack/pkg/store"
	uuid "github.com/nu7hatch/gouuid"
)

type Ctx struct {
	id         string
	store      *store.Service
	session    *session.Session
	req        *http.Request
	wr         http.ResponseWriter
	handlers   []Handler
	next       int
	statusCode int
}

func New(wr http.ResponseWriter, req *http.Request, handlers ...Handler) Ctx {
	// not allowing nil
	for i, h := range handlers {
		if h == nil {
			handlers = slices.Delete(handlers, i, 1)
		}
	}

	var id string
	u, err := uuid.NewV4()
	if err == nil {
		id = u.String()
	} else {
		id = "unknown-uuid"
	}

	return Ctx{
		id:         id,
		wr:         wr,
		req:        req,
		store:      store.New(),
		session:    session.New(),
		handlers:   handlers,
		statusCode: http.StatusOK,
	}
}

// current request context
func (c Ctx) Context() context.Context {
	return c.req.Context()
}

func (c Ctx) Store() *store.Service {
	return c.store
}

func (c Ctx) Next() error {
	if c.next >= len(c.handlers) {
		return nil
	}
	c.next++

	return c.handlers[c.next-1](c)
}

func (c Ctx) ID() string {
	return c.id
}

func (c Ctx) Status() int {
	return c.statusCode
}

func (c Ctx) SetCurrentAccount(a adapter.Account) {
	c.store.Set("current_account", a)
}

func (c Ctx) GetCurrentAccount() adapter.Account {
	var a adapter.Account
	_ = c.store.Get("current_account", &a)

	return a
}
