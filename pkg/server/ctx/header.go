package ctx

import (
	"net/textproto"

	"git.martianoids.com/martianoids/martian-stack/pkg/server/web"
)

func (c Ctx) SetHeader(key, value string) {
	c.wr.Header().Add(key, value)
}

func (c Ctx) GetRequestHeader(key string) string {
	return c.req.Header.Get(textproto.CanonicalMIMEHeaderKey(key))
}

func (c Ctx) Accept() string {
	return c.GetRequestHeader(web.HeaderAccept)
}

func (c Ctx) AcceptsJSON() bool {
	return c.GetRequestHeader(web.HeaderAccept) == web.MIMEApplicationJSON
}

func (c Ctx) AcceptsHTML() bool {
	return c.GetRequestHeader(web.HeaderAccept) == web.MIMETextHTML
}

func (c Ctx) AcceptsPlainText() bool {
	return c.GetRequestHeader(web.HeaderAccept) == web.MIMETextPlain
}
