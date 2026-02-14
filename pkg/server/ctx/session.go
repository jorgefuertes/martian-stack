package ctx

import "github.com/jorgefuertes/martian-stack/pkg/server/session"

func (c Ctx) Session() *session.Session {
	return c.session
}
