package helper_test

import (
	"testing"

	"github.com/jorgefuertes/martian-stack/pkg/helper"
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
			if got != tc.want {
				t.Errorf("replacePathParams(%q) = %q, want %q", tc.path, got, tc.want)
			}
		})
	}
}
