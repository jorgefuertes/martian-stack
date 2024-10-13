package ctx

import "git.martianoids.com/martianoids/martian-stack/pkg/server/session"

func (c Ctx) Session() *session.Session {
	return c.session
}
