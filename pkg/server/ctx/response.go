package ctx

import (
	"bytes"
	"encoding/json"
	"fmt"
	"mime"
	"time"

	"git.martianoids.com/martianoids/martian-stack/pkg/server/web"
)

func (c Ctx) SetContentType(contentType string) {
	c.SetHeader(web.HeaderContentType, contentType)
}

func (c Ctx) WithHeader(key, value string) Ctx {
	c.SetHeader(key, value)

	return c
}

func (c Ctx) SetCookie(name, value string, expire time.Duration) {
	c.SetHeader(web.HeaderSetCookie, fmt.Sprintf("%s=%s; Max-Age=%0f; Path=/; Domain=%s;",
		name, value, expire.Seconds(), c.req.Host))
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
	return c.WithHeader(web.HeaderContentType, web.MIMETextHTML).Write([]byte(s))
}

// set content-type as text/plain and write the string
// set status to http.StatusOK if no prior code is set
func (c Ctx) SendString(s string) error {
	return c.WithHeader(web.HeaderContentType, web.MIMETextPlain).Write([]byte(s))
}

// set content-type as application/html and write marshalled object as json string
// set status to http.StatusOK if no prior code is set
func (c Ctx) SendJSON(obj any) error {
	c.SetHeader(web.HeaderContentType, web.MIMEApplicationJSON)
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
	c.SetHeader(web.HeaderContentType, mime.TypeByExtension(filename))
	c.SetHeader(web.HeaderContentDisposition, "attachment; filename="+filename)

	return c.Write(contents.Bytes())
}

func (c Ctx) Write(b []byte) error {
	_, err := c.wr.Write(b)

	return err
}
