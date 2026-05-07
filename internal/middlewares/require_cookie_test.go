package middlewares

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"avgys-gophermat/internal/service/auth"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestRequireCookie_MissingCookie(t *testing.T) {
	called := false
	h := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		called = true
	})

	req := httptest.NewRequest(http.MethodGet, "http://example.com", nil)
	recorder := httptest.NewRecorder()

	RequireCookie(h).ServeHTTP(recorder, req)

	assert.False(t, called)
	assert.Equal(t, http.StatusUnauthorized, recorder.Code)
}

func TestRequireCookie_InvalidToken(t *testing.T) {
	called := false
	h := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		called = true
	})

	req := httptest.NewRequest(http.MethodGet, "http://example.com", nil)
	req.AddCookie(&http.Cookie{Name: auth.CookieName, Value: "bad"})
	recorder := httptest.NewRecorder()

	RequireCookie(h).ServeHTTP(recorder, req)

	assert.False(t, called)
	assert.Equal(t, http.StatusUnauthorized, recorder.Code)
}

func TestRequireCookie_ValidToken(t *testing.T) {
	claims := auth.NewToken(1, "user")
	tokenStr, err := claims.ToString()
	require.NoError(t, err)

	h := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		ctxClaims, err := auth.GetFromContext(r.Context())
		require.NoError(t, err)
		require.Equal(t, int64(1), ctxClaims.UserID)
		w.WriteHeader(http.StatusOK)
	})

	req := httptest.NewRequest(http.MethodGet, "http://example.com", nil)
	req.AddCookie(&http.Cookie{Name: auth.CookieName, Value: tokenStr})
	recorder := httptest.NewRecorder()

	RequireCookie(h).ServeHTTP(recorder, req)

	assert.Equal(t, http.StatusOK, recorder.Code)

	cookies := recorder.Result().Cookies()
	found := false
	for _, cookie := range cookies {
		if cookie.Name == auth.CookieName {
			found = true
			break
		}
	}
	assert.True(t, found)
}
