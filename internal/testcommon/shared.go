package testcommon

import (
	"net/http"
	"testing"

	"github.com/stretchr/testify/assert"
)

type ResponseWant struct {
	StatusCode int
	Body       string
	Headers    map[string]string
}

func CheckResponseFields(t *testing.T, res *http.Response, resBody []byte, want ResponseWant) {

	assert.Equal(t, want.StatusCode, res.StatusCode)

	contentType := res.Header.Get("Content-Type")

	if contentType == "application/json" {
		assert.JSONEq(t, want.Body, string(resBody))
	} else {
		assert.Equal(t, want.Body, string(resBody))
	}

	if want.Headers != nil {
		for k, v := range want.Headers {
			assert.Equal(t, res.Header.Get(k), v)
		}
	}
}
