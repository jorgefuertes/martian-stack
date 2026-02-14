package middleware

import (
	"fmt"
	"net/http"

	"github.com/jorgefuertes/martian-stack/pkg/server/ctx"
	"github.com/jorgefuertes/martian-stack/pkg/server/servererror"
)

// NewRecovery returns a middleware that recovers from panics in downstream
// handlers and converts them into a 500 Internal Server Error response.
func NewRecovery() ctx.Handler {
	return func(c ctx.Ctx) (err error) {
		defer func() {
			if r := recover(); r != nil {
				err = servererror.Error{
					Code: http.StatusInternalServerError,
					Msg:  fmt.Sprintf("panic: %v", r),
				}
			}
		}()

		return c.Next()
	}
}
