package ctx

import (
	"bytes"
	"encoding/json"
	"mime"
	"net/http"
	"strings"
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
	cookie := &http.Cookie{
		Name:     name,
		Value:    value,
		Path:     "/",
		Domain:   c.req.Host,
		MaxAge:   int(expire.Seconds()),
		HttpOnly: true,
		Secure:   c.req.TLS != nil || strings.EqualFold(c.req.Header.Get("X-Forwarded-Proto"), "https"),
		SameSite: http.SameSiteLaxMode,
	}

	// strip port from domain if present
	if idx := strings.IndexByte(cookie.Domain, ':'); idx != -1 {
		cookie.Domain = cookie.Domain[:idx]
	}

	http.SetCookie(c.wr, cookie)
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
	// sanitize filename: remove path separators and quote for header safety
	sanitized := strings.ReplaceAll(filename, "\\", "")
	sanitized = strings.ReplaceAll(sanitized, "/", "")
	sanitized = strings.ReplaceAll(sanitized, "\"", "")

	c.SetHeader(web.HeaderContentType, mime.TypeByExtension(filename))
	c.SetHeader(web.HeaderContentDisposition, "attachment; filename=\""+sanitized+"\"")

	return c.Write(contents.Bytes())
}

func (c Ctx) Write(b []byte) error {
	_, err := c.wr.Write(b)

	return err
}
