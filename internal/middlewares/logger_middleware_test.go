package middlewares

import (
	"net/http"
	"net/http/httptest"
)

func (s *MiddlewaresSuite) TestWrapWriter_TracksStatusAndSize() {
	recorder := httptest.NewRecorder()
	wrapped := wrapWriter(recorder)

	wrapped.WriteHeader(http.StatusCreated)
	_, err := wrapped.Write([]byte("hello"))
	s.Require().NoError(err)

	s.Equal(http.StatusCreated, wrapped.logData.statusCode)
	s.Equal(len("hello"), wrapped.logData.responseSize)
}

func (s *MiddlewaresSuite) TestWithLogging_Passthrough() {
	h := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte("ok"))
	})

	recorder := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "http://example.com", nil)

	WithLogging(h).ServeHTTP(recorder, req)

	s.Equal(http.StatusOK, recorder.Code)
	s.Equal("ok", recorder.Body.String())
}
