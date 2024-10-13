package server_test

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"
	"testing"
	"time"

	"git.martianoids.com/martianoids/martian-stack/pkg/server/web"
	"github.com/stretchr/testify/require"
)

func composeURL(path string) string {
	if !strings.HasPrefix(path, "/") {
		path = "/" + path
	}

	return fmt.Sprintf("http://%s:%s%s", host, port, path)
}

func call(
	method web.Method,
	acceptContetType string,
	cookies []*http.Cookie,
	path string,
	obj any,
) (*http.Response, error) {
	var req *http.Request
	var err error
	if obj != nil {
		b, _ := json.Marshal(obj)
		reqBodyReader := bytes.NewReader(b)
		req, err = http.NewRequest(method.String(), composeURL(path), reqBodyReader)
		req.Header.Set(web.HeaderContentType, "application/json")
	} else {
		req, err = http.NewRequest(method.String(), composeURL(path), nil)
	}
	if err != nil {
		return nil, err
	}

	if acceptContetType != "" {
		req.Header.Set(web.HeaderAccept, acceptContetType)
	}

	for _, c := range cookies {
		req.AddCookie(c)
	}

	client := &http.Client{Timeout: timeoutSeconds * time.Second}
	return client.Do(req)
}

func bodyAsString(t *testing.T, res *http.Response) string {
	t.Helper()

	b, err := io.ReadAll(res.Body)
	require.NoError(t, err)
	return string(b)
}

func bodyAsJSON(t *testing.T, res *http.Response, dest any) {
	t.Helper()

	b, err := io.ReadAll(res.Body)
	require.NoError(t, err, "Body: %s", string(b))

	err = json.Unmarshal(b, dest)
	require.NoError(t, err)
}
