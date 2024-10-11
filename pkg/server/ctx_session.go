package server

import "git.martianoids.com/martianoids/martian-stack/pkg/server/session"


func (c Ctx) Session() *session.Session {
	return c.session
}

func (c *Ctx) SetSession(s *session.Session) {
	c.session = s
}
