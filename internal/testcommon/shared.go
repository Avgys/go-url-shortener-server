package testcommon

import (
	"io"
	"net/http"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

type ResponseWant struct {
	StatusCode int
	Body       string
	Headers    map[string]string
}

func CheckResponseFields(t *testing.T, res *http.Response, want ResponseWant) {

	resBody, err := io.ReadAll(res.Body)
	require.NoError(t, err)

	assert.Equal(t, want.StatusCode, res.StatusCode)

	contentType := res.Header.Get("Content-Type")

	if want.Body != "" {
		if contentType == "application/json" {
			assert.JSONEq(t, want.Body, string(resBody))
		} else {
			assert.Equal(t, want.Body, strings.TrimRight(string(resBody), "\n"))
		}
	}

	if want.Headers != nil {
		for k, v := range want.Headers {
			assert.Equal(t, res.Header.Get(k), v)
		}
	}
}
