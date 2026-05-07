package middlewares

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestWrapWriter_TracksStatusAndSize(t *testing.T) {
	recorder := httptest.NewRecorder()
	wrapped := wrapWriter(recorder)

	wrapped.WriteHeader(http.StatusCreated)
	_, err := wrapped.Write([]byte("hello"))
	require.NoError(t, err)

	assert.Equal(t, http.StatusCreated, wrapped.logData.statusCode)
	assert.Equal(t, len("hello"), wrapped.logData.responseSize)
}

func TestWithLogging_Passthrough(t *testing.T) {
	h := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte("ok"))
	})

	recorder := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "http://example.com", nil)

	WithLogging(h).ServeHTTP(recorder, req)

	assert.Equal(t, http.StatusOK, recorder.Code)
	assert.Equal(t, "ok", recorder.Body.String())
}
