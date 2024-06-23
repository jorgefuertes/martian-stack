package server

import (
	"bytes"
	"context"
	"encoding/json"
	"mime"
	"net/http"
	"net/textproto"
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

func (c Ctx) setHeader(key, value string) {
	c.wr.Header().Add(key, value)
}

func (c Ctx) GetRequestHeader(key string) string {
	return c.req.Header.Get(textproto.CanonicalMIMEHeaderKey(key))
}

func (c Ctx) Accept() string {
	return c.GetRequestHeader(HeaderAccept)
}

func (c Ctx) AcceptsJSON() bool {
	return c.GetRequestHeader(HeaderAccept) == MIMEApplicationJSON
}

func (c Ctx) SetContentType(contentType string) {
	c.setHeader(HeaderContentType, contentType)
}

func (c Ctx) WithHeader(key, value string) Ctx {
	c.setHeader(key, value)

	return c
}

// explicit status code, set it before any write
func (c Ctx) WithStatus(code int) Ctx {
	c.wr.WriteHeader(code)

	return c
}

// set content-type as text/html and write the html string
// set status to http.StatusOK if no prior code is set
func (c Ctx) SendHtml(s string) error {
	return c.WithHeader(HeaderContentType, MIMETextHTML).Write([]byte(s))
}

// set content-type as text/plain and write the string
// set status to http.StatusOK if no prior code is set
func (c Ctx) SendString(s string) error {
	return c.WithHeader(HeaderContentType, MIMETextPlain).Write([]byte(s))
}

// set content-type as application/html and write marshalled object as json string
// set status to http.StatusOK if no prior code is set
func (c Ctx) SendJSON(obj any) error {
	c.setHeader(HeaderContentType, MIMEApplicationJSON)
	b, err := json.Marshal(obj)
	if err != nil {
		return err
	}

	return c.Write(b)
}

// Content-type: filename extension mime type
// Content-Disposition: attachment; filename="logo.png"
// Status: http.StatusOK if no prior code is set
func (c Ctx) SendAttachment(filename string, contents *bytes.Buffer) error {
	c.setHeader(HeaderContentType, mime.TypeByExtension(filename))
	c.setHeader(HeaderContentDisposition, "attachment; filename="+filename)

	return c.Write(contents.Bytes())
}

func (c Ctx) Write(b []byte) error {
	_, err := c.wr.Write(b)

	return err
}
