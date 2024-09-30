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

	"git.martianoids.com/martianoids/martian-stack/pkg/httpconst"
	"github.com/stretchr/testify/require"
)

func composeURL(path string) string {
	if !strings.HasPrefix(path, "/") {
		path = "/" + path
	}

	return fmt.Sprintf("http://%s:%s%s", host, port, path)
}

func call(method httpconst.Method, acceptContetType string, path string, obj any) (*http.Response, error) {
	var req *http.Request
	var err error
	if obj != nil {
		b, _ := json.Marshal(obj)
		reqBodyReader := bytes.NewReader(b)
		req, err = http.NewRequest(method.String(), composeURL(path), reqBodyReader)
		req.Header.Set(httpconst.HeaderContentType, "application/json")
	} else {
		req, err = http.NewRequest(method.String(), composeURL(path), nil)
	}
	if err != nil {
		return nil, err
	}

	if acceptContetType != "" {
		req.Header.Set(httpconst.HeaderAccept, acceptContetType)
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
