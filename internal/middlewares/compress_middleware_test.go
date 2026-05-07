package middlewares

import (
	"bytes"
	"compress/gzip"
	"io"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func gzipPayload(t *testing.T, data string) []byte {
	t.Helper()

	var buf bytes.Buffer
	gz := gzip.NewWriter(&buf)
	_, err := gz.Write([]byte(data))
	require.NoError(t, err)
	require.NoError(t, gz.Close())

	return buf.Bytes()
}

func TestWithCompression_GzipRequest(t *testing.T) {
	body := gzipPayload(t, "hello")
	req := httptest.NewRequest(http.MethodPost, "http://example.com", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Content-Encoding", "gzip")

	var gotBody string
	h := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		data, err := io.ReadAll(r.Body)
		require.NoError(t, err)
		gotBody = string(data)
		_, _ = w.Write([]byte("resp"))
	})

	recorder := httptest.NewRecorder()
	WithCompression(h).ServeHTTP(recorder, req)

	assert.Equal(t, "hello", gotBody)
	assert.Equal(t, "resp", recorder.Body.String())
}

func TestWithCompression_DecodeError(t *testing.T) {
	req := httptest.NewRequest(http.MethodPost, "http://example.com", bytes.NewReader([]byte("bad gzip")))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Content-Encoding", "gzip")

	called := false
	h := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		called = true
	})

	recorder := httptest.NewRecorder()
	WithCompression(h).ServeHTTP(recorder, req)

	assert.False(t, called)
	assert.Equal(t, http.StatusInternalServerError, recorder.Code)
}
