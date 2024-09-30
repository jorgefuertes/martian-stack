package server

import (
	"bytes"
	"context"
	"encoding/json"
	"io"
	"mime"
	"net/http"
	"net/textproto"
	"net/url"
	"strings"

	"git.martianoids.com/martianoids/martian-stack/pkg/httpconst"
	uuid "github.com/nu7hatch/gouuid"
)

type Ctx struct {
	id         string
	store      *store
	req        *http.Request
	wr         http.ResponseWriter
	handlers   []Handler
	next       int
	statusCode int
}

func NewCtx(wr http.ResponseWriter, req *http.Request, handlers ...Handler) Ctx {
	var id string
	u, err := uuid.NewV4()
	if err == nil {
		id = u.String()
	} else {
		id = "unknown-uuid"
	}

	return Ctx{id: id, wr: wr, req: req, store: newStore(), handlers: handlers, statusCode: http.StatusOK}
}

func (c Ctx) Context() context.Context {
	return c.req.Context()
}

func (c Ctx) Store() *store {
	return c.store
}

func (c Ctx) Next() error {
	if c.next >= len(c.handlers) {
		return nil
	}
	c.next++

	return c.handlers[c.next-1](c)
}

func (c Ctx) SetHeader(key, value string) {
	c.wr.Header().Add(key, value)
}

func (c Ctx) GetRequestHeader(key string) string {
	return c.req.Header.Get(textproto.CanonicalMIMEHeaderKey(key))
}

func (c Ctx) ID() string {
	return c.id
}

func (c Ctx) Method() string {
	return c.req.Method
}

func (c Ctx) Path() string {
	return c.req.RequestURI
}

func (c Ctx) UserIP() string {
	return strings.Split(c.req.RemoteAddr, ":")[0]
}

func (c Ctx) Status() int {
	return c.statusCode
}

func (c Ctx) Accept() string {
	return c.GetRequestHeader(httpconst.HeaderAccept)
}

func (c Ctx) AcceptsJSON() bool {
	return c.GetRequestHeader(httpconst.HeaderAccept) == httpconst.MIMEApplicationJSON
}

func (c Ctx) AcceptsHTML() bool {
	return c.GetRequestHeader(httpconst.HeaderAccept) == httpconst.MIMETextHTML
}

func (c Ctx) AcceptsPlainText() bool {
	return c.GetRequestHeader(httpconst.HeaderAccept) == httpconst.MIMETextPlain
}

func (c Ctx) SetContentType(contentType string) {
	c.SetHeader(httpconst.HeaderContentType, contentType)
}

func (c Ctx) WithHeader(key, value string) Ctx {
	c.SetHeader(key, value)

	return c
}

// explicit status code, set it before any write
func (c Ctx) WithStatus(code int) Ctx {
	c.statusCode = code
	c.wr.WriteHeader(code)

	return c
}

// set content-type as text/html and write the html string
// set status to http.StatusOK if no prior code is set
func (c Ctx) SendHTML(s string) error {
	return c.WithHeader(httpconst.HeaderContentType, httpconst.MIMETextHTML).Write([]byte(s))
}

// set content-type as text/plain and write the string
// set status to http.StatusOK if no prior code is set
func (c Ctx) SendString(s string) error {
	return c.WithHeader(httpconst.HeaderContentType, httpconst.MIMETextPlain).Write([]byte(s))
}

// set content-type as application/html and write marshalled object as json string
// set status to http.StatusOK if no prior code is set
func (c Ctx) SendJSON(obj any) error {
	c.SetHeader(httpconst.HeaderContentType, httpconst.MIMEApplicationJSON)
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
	c.SetHeader(httpconst.HeaderContentType, mime.TypeByExtension(filename))
	c.SetHeader(httpconst.HeaderContentDisposition, "attachment; filename="+filename)

	return c.Write(contents.Bytes())
}

func (c Ctx) Write(b []byte) error {
	_, err := c.wr.Write(b)

	return err
}

// helper to compose and HttpError to be used as error return
func (c Ctx) Error(code int, message any) HttpError {
	var msg string

	switch m := message.(type) {
	case string:
		msg = m
	case error:
		msg = m.Error()
	default:
		msg = http.StatusText(code)
	}

	return HttpError{Code: code, Msg: msg}
}

func (c Ctx) Param(key string) string {
	value := stringOrString(c.req.PathValue(key), c.req.URL.Query().Get(key))
	// decode url encoded parameters
	if strings.Contains(value, "%") {
		decoded, err := url.QueryUnescape(value)
		if err == nil {
			value = decoded
		}
	}

	return value
}

// unmarshal the request body into dest
func (c Ctx) UnmarshalBody(dest any) error {
	return json.NewDecoder(c.req.Body).Decode(dest)
}

type Component interface {
	Render(ctx context.Context, wr io.Writer) error
}

func (c Ctx) Render(f Component) error {
	return f.Render(c.Context(), c.wr)
}
