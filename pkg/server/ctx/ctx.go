package ctx

import (
	"context"
	"net/http"

	"git.martianoids.com/martianoids/martian-stack/pkg/server/adapter"
	"git.martianoids.com/martianoids/martian-stack/pkg/server/session"
	"git.martianoids.com/martianoids/martian-stack/pkg/store"
	"github.com/google/uuid"
)

// state holds mutable fields shared across Ctx copies.
// Since Ctx is passed by value through the handler chain,
// this pointer ensures mutations (e.g. status code set by a handler)
// are visible to middleware that runs after (e.g. logging).
type state struct {
	next       int
	statusCode int
}

type Ctx struct {
	id       string
	store    *store.Service
	session  *session.Session
	req      *http.Request
	wr       http.ResponseWriter
	handlers []Handler
	state    *state
}

func New(wr http.ResponseWriter, req *http.Request, handlers ...Handler) Ctx {
	// filter out nil handlers
	n := 0
	for _, h := range handlers {
		if h != nil {
			handlers[n] = h
			n++
		}
	}
	handlers = handlers[:n]

	id := uuid.New().String()

	return Ctx{
		id:       id,
		wr:       wr,
		req:      req,
		store:    store.New(),
		session:  session.New(),
		handlers: handlers,
		state:    &state{statusCode: http.StatusOK},
	}
}

// Context returns the request context.
func (c Ctx) Context() context.Context {
	return c.req.Context()
}

// WithContext returns a copy of Ctx with the request context replaced.
// Use this for deadlines, cancellation or passing values to downstream handlers.
func (c Ctx) WithContext(reqCtx context.Context) Ctx {
	c.req = c.req.WithContext(reqCtx)
	return c
}

func (c Ctx) Store() *store.Service {
	return c.store
}

func (c Ctx) Next() error {
	if c.state.next >= len(c.handlers) {
		return nil
	}
	c.state.next++

	return c.handlers[c.state.next-1](c)
}

func (c Ctx) ID() string {
	return c.id
}

func (c Ctx) Status() int {
	return c.state.statusCode
}

func (c Ctx) SetCurrentAccount(a adapter.Account) {
	c.store.Set("current_account", a)
}

func (c Ctx) GetCurrentAccount() adapter.Account {
	var a adapter.Account
	_ = c.store.Get("current_account", &a)

	return a
}
