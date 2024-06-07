package server

import (
	"context"
	"net/http"
)

type Ctx struct {
	store    *store
	req      *http.Request
	wr       http.ResponseWriter
	handlers []Handler
	next     int
}

func newCtx(wr http.ResponseWriter, req *http.Request, handlers ...Handler) Ctx {
	return Ctx{wr: wr, req: req, store: newStore(), handlers: handlers}
}

func (c Ctx) Context() context.Context {
	return c.req.Context()
}

func (c Ctx) Next() error {
	if c.next >= len(c.handlers) {
		return nil
	}
	c.next++

	return c.handlers[c.next-1](c)
}

func (c *Ctx) SetStatus(code int) *Ctx {
	c.wr.WriteHeader(code)

	return c
}

func (c *Ctx) SendString(s string) error {
	_, err := c.wr.Write([]byte(s))

	return err
}
