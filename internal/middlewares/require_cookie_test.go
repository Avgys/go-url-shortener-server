package middlewares

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/suite"
	"go-url-shortener/internal/service/auth"
)

func TestMiddlewaresSuite(t *testing.T) {
	suite.Run(t, new(MiddlewaresSuite))
}

type MiddlewaresSuite struct {
	suite.Suite
}

func (s *MiddlewaresSuite) TestRequireCookie_MissingCookie() {
	called := false
	h := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		called = true
	})

	req := httptest.NewRequest(http.MethodGet, "http://example.com", nil)
	recorder := httptest.NewRecorder()

	AuthRequireCookie(h).ServeHTTP(recorder, req)

	s.False(called)
	s.Equal(http.StatusUnauthorized, recorder.Code)
}

func (s *MiddlewaresSuite) TestRequireCookie_InvalidToken() {
	called := false
	h := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		called = true
	})

	req := httptest.NewRequest(http.MethodGet, "http://example.com", nil)
	req.AddCookie(&http.Cookie{Name: auth.CookieName, Value: "bad"})
	recorder := httptest.NewRecorder()

	AuthRequireCookie(h).ServeHTTP(recorder, req)

	s.False(called)
	s.Equal(http.StatusUnauthorized, recorder.Code)
}

func (s *MiddlewaresSuite) TestRequireCookie_ValidToken() {
	claims := auth.NewToken(1, "user")
	tokenStr, err := claims.ToString()
	s.Require().NoError(err)

	h := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		ctxClaims, err := auth.GetFromContext(r.Context())
		s.Require().NoError(err)
		s.Equal(int64(1), ctxClaims.UserID)
		w.WriteHeader(http.StatusOK)
	})

	req := httptest.NewRequest(http.MethodGet, "http://example.com", nil)
	req.AddCookie(&http.Cookie{Name: auth.CookieName, Value: tokenStr})
	recorder := httptest.NewRecorder()

	AuthRequireCookie(h).ServeHTTP(recorder, req)

	s.Equal(http.StatusOK, recorder.Code)

	cookies := recorder.Result().Cookies()
	found := false
	for _, cookie := range cookies {
		if cookie.Name == auth.CookieName {
			found = true
			break
		}
	}
	s.True(found)
}
