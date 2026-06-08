package middlewares

import (
	"bytes"
	"compress/gzip"
	"io"
	"net/http"
	"net/http/httptest"
)

func (s *MiddlewaresSuite) gzipPayload(data string) []byte {
	var buf bytes.Buffer
	gz := gzip.NewWriter(&buf)
	_, err := gz.Write([]byte(data))
	s.Require().NoError(err)
	s.Require().NoError(gz.Close())

	return buf.Bytes()
}

func (s *MiddlewaresSuite) TestWithCompression_GzipRequest() {
	body := s.gzipPayload("hello")
	req := httptest.NewRequest(http.MethodPost, "http://example.com", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Content-Encoding", "gzip")

	var gotBody string
	h := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		data, err := io.ReadAll(r.Body)
		s.Require().NoError(err)
		gotBody = string(data)
		_, _ = w.Write([]byte("resp"))
	})

	recorder := httptest.NewRecorder()
	WithCompression(h).ServeHTTP(recorder, req)

	s.Equal("hello", gotBody)
	s.Equal("resp", recorder.Body.String())
}

func (s *MiddlewaresSuite) TestWithCompression_DecodeError() {
	req := httptest.NewRequest(http.MethodPost, "http://example.com", bytes.NewReader([]byte("bad gzip")))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Content-Encoding", "gzip")

	called := false
	h := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		called = true
	})

	recorder := httptest.NewRecorder()
	WithCompression(h).ServeHTTP(recorder, req)

	s.False(called)
	s.Equal(http.StatusInternalServerError, recorder.Code)
}
