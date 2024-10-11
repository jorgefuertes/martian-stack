package middleware

import (
	"time"

	"git.martianoids.com/martianoids/martian-stack/pkg/server"
	"git.martianoids.com/martianoids/martian-stack/pkg/server/session"
	"git.martianoids.com/martianoids/martian-stack/pkg/service/cache"
)

const cookieName = "martian_session_id"
const defaultExpiration = 48 * time.Hour

func NewSession(cacheSvc cache.Service) server.Handler {
	return func(c server.Ctx) error {
		id := c.GetCookie(cookieName)
		s := session.New()

		if id == "" {
			c.SetCookie(cookieName, s.ID, defaultExpiration)
		} else {
			s.ID = id
		}

		if cacheSvc.Exists(c.Context(), s.KeyID()) {
			// recover session from cache
			b, err := cacheSvc.GetBytes(c.Context(), s.KeyID())
			if err != nil {
				return err
			}

			if err := s.UnmarshalJSON(b); err != nil {
				return err
			}
		}

		c.SetSession(s)
		hErr := c.Next()

		// save session to cache if dirty
		if s.Data().IsDirty() {
			if err := cacheSvc.Set(c.Context(), s.KeyID(), s.Data(), defaultExpiration); err != nil {
				return err
			}
		}

		return hErr
	}
}
