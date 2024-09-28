package middleware

import (
	"crypto/subtle"
	"encoding/base64"
	"net/http"
	"regexp"
	"strings"

	"git.martianoids.com/martianoids/martian-stack/pkg/service/server"
	"git.martianoids.com/martianoids/martian-stack/pkg/service/server/web"
)

func NewBasicAuthMw(username, password string) server.Handler {
	return func(c server.Ctx) error {
		auth := c.GetRequestHeader(web.HeaderAuthorization)

		// check auth header
		if auth == "" {
			c.SetHeader(web.HeaderWWWAuthenticate, "Basic realm=\"Restricted\"")
			return c.Error(http.StatusUnauthorized, nil)
		}

		r := regexp.MustCompile(`^Basic ([a-zA-Z0-9+/=]*)$`)
		if !r.MatchString(auth) {
			c.SetHeader(web.HeaderWWWAuthenticate, "Basic realm=\"Restricted\"")
			return c.Error(http.StatusBadRequest, nil)
		}

		// decode
		payload := r.FindStringSubmatch(auth)
		if len(payload) != 2 {
			c.SetHeader(web.HeaderWWWAuthenticate, "Basic realm=\"Restricted\"")
			return c.Error(http.StatusBadRequest, nil)
		}

		bToken, err := base64.StdEncoding.DecodeString(payload[1])
		if err != nil {
			c.SetHeader(web.HeaderWWWAuthenticate, "Basic realm=\"Restricted\"")
			return c.Error(http.StatusBadRequest, nil)
		}

		// get username and password
		pair := strings.SplitN(string(bToken), ":", 2)
		if len(pair) != 2 {
			c.SetHeader(web.HeaderWWWAuthenticate, "Basic realm=\"Restricted\"")
			return c.Error(http.StatusBadRequest, nil)
		}

		// time constant compare
		if subtle.ConstantTimeCompare([]byte(pair[0]), []byte(username)) != 1 ||
			subtle.ConstantTimeCompare([]byte(pair[1]), []byte(password)) != 1 {
			c.SetHeader(web.HeaderWWWAuthenticate, "Basic realm=\"Restricted\"")
			return c.Error(http.StatusUnauthorized, nil)
		}

		// auth OK
		return nil
	}
}
