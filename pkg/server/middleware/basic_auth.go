package middleware

import (
	"crypto/sha256"
	"crypto/subtle"
	"encoding/base64"
	"net/http"
	"regexp"
	"strings"

	"git.martianoids.com/martianoids/martian-stack/pkg/server/ctx"
	"git.martianoids.com/martianoids/martian-stack/pkg/server/web"
)

var basicAuthRegexp = regexp.MustCompile(`^Basic ([a-zA-Z0-9+/=]*)$`)

func NewBasicAuth(username, password string) ctx.Handler {
	// pre-compute expected hashes to prevent length-based timing leaks
	expectedUserHash := sha256.Sum256([]byte(username))
	expectedPassHash := sha256.Sum256([]byte(password))

	return func(c ctx.Ctx) error {
		auth := c.GetRequestHeader(web.HeaderAuthorization)

		// check auth header
		if auth == "" {
			c.SetHeader(web.HeaderWWWAuthenticate, "Basic realm=\"Restricted\"")
			return c.Error(http.StatusUnauthorized, nil)
		}

		if !basicAuthRegexp.MatchString(auth) {
			c.SetHeader(web.HeaderWWWAuthenticate, "Basic realm=\"Restricted\"")
			return c.Error(http.StatusBadRequest, nil)
		}

		// decode
		payload := basicAuthRegexp.FindStringSubmatch(auth)
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

		// constant-time compare using SHA256 hashes to prevent length-based timing leaks
		givenUserHash := sha256.Sum256([]byte(pair[0]))
		givenPassHash := sha256.Sum256([]byte(pair[1]))

		userMatch := subtle.ConstantTimeCompare(givenUserHash[:], expectedUserHash[:])
		passMatch := subtle.ConstantTimeCompare(givenPassHash[:], expectedPassHash[:])

		if userMatch != 1 || passMatch != 1 {
			c.SetHeader(web.HeaderWWWAuthenticate, "Basic realm=\"Restricted\"")
			return c.Error(http.StatusUnauthorized, nil)
		}

		// auth OK
		return c.Next()
	}
}
