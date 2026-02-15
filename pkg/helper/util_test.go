package helper_test

import (
	"testing"

	"github.com/jorgefuertes/martian-stack/pkg/helper"
	"github.com/stretchr/testify/assert"
)

func TestReplacePathParams(t *testing.T) {
	testCases := []struct {
		path string
		want string
	}{
		{"/hello", "/hello"},
		{"/hello/:name", "/hello/{name}"},
		{"/:name", "/{name}"},
		{"/:name/:age", "/{name}/{age}"},
		{"/:name/:age/:city", "/{name}/{age}/{city}"},
		{"/:name/:age/:city/:id", "/{name}/{age}/{city}/{id}"},
		{"/:name/:age/:city/:id/:extra", "/{name}/{age}/{city}/{id}/{extra}"},
		{"/:name/:age/:city/:id/:extra/:extra2", "/{name}/{age}/{city}/{id}/{extra}/{extra2}"},
	}

	for _, tc := range testCases {
		t.Run(tc.path, func(t *testing.T) {
			got := helper.ReplacePathParams(tc.path)
			assert.Equal(t, tc.want, got)
		})
	}
}
