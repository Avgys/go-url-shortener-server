package auth_middlewares

import (
	"context"
	"net/http"
	"time"

	"github.com/Avgys/go-url-shortener-server/internal/auth"
	"github.com/Avgys/go-url-shortener-server/internal/auth/jwt_token"
)

func RequireCookie(h http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {

		authCookie, err := r.Cookie(auth.AUTH_COOKIE)

		if err == http.ErrNoCookie {
			w.WriteHeader(http.StatusUnauthorized)
			return
		}

		claims, err := jwt_token.ParseToken(authCookie.Value)

		if claims.UserID == 0 || err != nil {
			w.WriteHeader(http.StatusUnauthorized)
			return
		}

		// refresh cookie expire time
		newCookie := &http.Cookie{Name: auth.AUTH_COOKIE, Value: authCookie.Value, Expires: time.Now().Add(jwt_token.TOKEN_EXP)}

		authCtx := context.WithValue(r.Context(), auth.CLAIMS, claims)
		r = r.WithContext(authCtx)

		http.SetCookie(w, newCookie)

		h.ServeHTTP(w, r)
	})
}
