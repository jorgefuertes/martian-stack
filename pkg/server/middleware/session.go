package middleware

import (
	"errors"
	"time"

	"git.martianoids.com/martianoids/martian-stack/pkg/server/ctx"
	"git.martianoids.com/martianoids/martian-stack/pkg/service/cache"
)

const (
	cookieName        = "martian_session_id"
	defaultExpiration = 48 * time.Hour

	SessionAutoStart   bool = true
	NoSessionAutoStart bool = false
)

func NewSession(cacheSvc cache.Service, autostart bool) ctx.Handler {
	return func(c ctx.Ctx) error {
		if autostart {
			if err := StartSession(c, cacheSvc); err != nil {
				return err
			}
		}

		hErr := c.Next()

		// save session to cache if dirty
		if c.Session().Data().IsDirty() {
			if err := cacheSvc.Set(c.Context(), c.Session().KeyID(), c.Session().Data(), defaultExpiration); err != nil {
				return errors.Join(err, hErr)
			}

			c.Session().Data().SetClean()
		}

		return hErr
	}
}

func StartSession(c ctx.Ctx, cacheSvc cache.Service) error {
	id := c.GetCookie(cookieName)

	if id == "" {
		c.SetCookie(cookieName, c.Session().ID, defaultExpiration)
	} else {
		c.Session().ID = id
	}

	if cacheSvc.Exists(c.Context(), c.Session().KeyID()) {
		// recover session from cache
		b, err := cacheSvc.GetBytes(c.Context(), c.Session().KeyID())
		if err != nil {
			return err
		}

		if err := c.Session().UnmarshalJSON(b); err != nil {
			return err
		}
	}

	return nil
}
