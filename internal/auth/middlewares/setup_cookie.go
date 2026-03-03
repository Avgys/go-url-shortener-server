package auth_middlewares

import (
	"context"
	"math/big"
	"net/http"
	"time"

	"crypto/rand"

	"github.com/Avgys/go-url-shortener-server/internal/auth"
	"github.com/Avgys/go-url-shortener-server/internal/auth/jwt_token"
	"github.com/Avgys/go-url-shortener-server/internal/logger"
)

func SetCookie(h http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {

		traceLogger := logger.Middleware(r.Context(), "SetCookie")

		authCookie, err := r.Cookie(auth.AUTH_COOKIE)

		resetCookie := false
		var claims *jwt_token.Claims

		if err == http.ErrNoCookie {
			resetCookie = true
		} else {
			claims, err = jwt_token.ParseToken(authCookie.Value)

			if err != nil || claims == nil || claims.UserID == 0 {
				resetCookie = true
			}
		}

		var tokenString string

		if resetCookie {
			userID, _ := rand.Int(rand.Reader, big.NewInt(1<<23))
			tokenString, claims, _ = jwt_token.NewTokenWithUserId(userID.Int64())

			traceLogger.Info().
				Str("Cookie", tokenString).
				Bool("IsNewCookie", true).
				Send()

		} else {
			tokenString = authCookie.Value

			traceLogger.Info().
				Str("Cookie", tokenString).
				Bool("IsNewCookie", false).
				Send()
		}

		authCtx := context.WithValue(r.Context(), auth.CLAIMS, claims)
		r = r.WithContext(authCtx)
		// refresh cookie expire time
		newCookie := &http.Cookie{Name: auth.AUTH_COOKIE, Value: tokenString, Expires: time.Now().Add(jwt_token.TOKEN_EXP)}
		http.SetCookie(w, newCookie)

		h.ServeHTTP(w, r)
	})
}
