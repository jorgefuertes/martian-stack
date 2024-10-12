package server

import "git.martianoids.com/martianoids/martian-stack/pkg/server/session"

func (c Ctx) Session() *session.Session {
	if c.session == nil {
		panic(ErrSessionNotStarted)
	}

	return c.session
}
