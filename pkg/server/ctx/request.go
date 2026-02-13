package ctx

import (
	"encoding/json"
	"net/http"
	"net/url"
	"strings"

	"git.martianoids.com/martianoids/martian-stack/pkg/helper"
)

func (c Ctx) Method() string {
	return c.req.Method
}

func (c Ctx) URL() *url.URL {
	return c.req.URL
}

func (c Ctx) Path() string {
	return c.req.URL.Path
}

func (c Ctx) UserIP() string {
	return strings.Split(c.req.RemoteAddr, ":")[0]
}

func (c Ctx) Param(key string) string {
	value := helper.StringOrString(c.req.PathValue(key), c.req.URL.Query().Get(key))
	// decode url encoded parameters
	if strings.Contains(value, "%") {
		decoded, err := url.QueryUnescape(value)
		if err == nil {
			value = decoded
		}
	}

	return value
}

func (c Ctx) GetCookie(name string) string {
	cookie, err := c.req.Cookie(name)
	if err != nil {
		return ""
	}

	return cookie.Value
}

// MaxBodySize is the default maximum request body size (1 MB)
const MaxBodySize int64 = 1 << 20

// unmarshal the request body into dest, limiting the body size to prevent abuse
func (c Ctx) UnmarshalBody(dest any) error {
	limited := http.MaxBytesReader(c.wr, c.req.Body, MaxBodySize)
	return json.NewDecoder(limited).Decode(dest)
}
