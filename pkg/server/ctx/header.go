package ctx

import (
	"net/textproto"
	"strings"

	"git.martianoids.com/martianoids/martian-stack/pkg/server/web"
)

func (c Ctx) SetHeader(key, value string) {
	c.wr.Header().Set(key, value)
}

// AddHeader appends a value to an existing header (multiple values for same key)
func (c Ctx) AddHeader(key, value string) {
	c.wr.Header().Add(key, value)
}

func (c Ctx) GetRequestHeader(key string) string {
	return c.req.Header.Get(textproto.CanonicalMIMEHeaderKey(key))
}

func (c Ctx) Accept() string {
	return c.GetRequestHeader(web.HeaderAccept)
}

func (c Ctx) AcceptsJSON() bool {
	return c.acceptsType(web.MIMEApplicationJSON)
}

func (c Ctx) AcceptsHTML() bool {
	return c.acceptsType(web.MIMETextHTML)
}

func (c Ctx) AcceptsPlainText() bool {
	return c.acceptsType(web.MIMETextPlain)
}

// acceptsType checks whether the Accept header includes the given MIME type
func (c Ctx) acceptsType(mimeType string) bool {
	accept := c.GetRequestHeader(web.HeaderAccept)
	if accept == "" {
		return false
	}

	for _, part := range strings.Split(accept, ",") {
		// strip quality parameter and whitespace: "text/html;q=0.9" -> "text/html"
		mediaType := strings.TrimSpace(strings.SplitN(part, ";", 2)[0])
		if mediaType == mimeType || mediaType == "*/*" {
			return true
		}
	}

	return false
}
