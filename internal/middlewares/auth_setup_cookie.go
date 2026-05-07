package middlewares

import (
	"math/big"
	"net/http"
	"time"

	"crypto/rand"

	"github.com/Avgys/go-url-shortener-server/internal/auth/jwttoken"
	"github.com/Avgys/go-url-shortener-server/internal/logger"
)

func SetCookie(h http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {

		traceLogger := logger.Middleware(r.Context(), "SetCookie")

		authCookie, err := r.Cookie(string(authCookieName))

		resetCookie := false
		var claims *jwttoken.Claims

		if err == http.ErrNoCookie {
			resetCookie = true
		} else {
			claims, err = jwttoken.ParseToken(authCookie.Value)

			if err != nil || claims == nil || claims.UserID == 0 {
				resetCookie = true
			}
		}

		var tokenString string

		if resetCookie {
			userID, err := rand.Int(rand.Reader, big.NewInt(1<<62))

			if err != nil {
				traceLogger.Error().Err(err).Msg("failed to generate random userID for auth cookie")
				http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
				return
			}

			tokenString, claims, err = jwttoken.NewTokenWithUserID(userID.Int64())

			if err != nil || claims == nil || tokenString == "" {
				if err != nil {
					traceLogger.Error().Err(err).Msg("failed to create JWT token for auth cookie")
				} else {
					traceLogger.Error().Msg("generated invalid JWT token or claims for auth cookie")
				}
				http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
				return
			}

			traceLogger.Info().
				Bool("IsNewCookie", true).
				Send()

		} else {
			tokenString = authCookie.Value

			traceLogger.Info().
				Bool("IsNewCookie", false).
				Send()
		}

		authCtx := claims.WithContext(r.Context())
		r = r.WithContext(authCtx)

		// refresh cookie expire time
		newCookie := createCookie(tokenString)
		http.SetCookie(w, newCookie)

		h.ServeHTTP(w, r)
	})
}

func createCookie(cookieValue string) *http.Cookie {
	newCookie := &http.Cookie{Name: string(authCookieName), Value: cookieValue, Expires: time.Now().Add(jwttoken.TokenExp), HttpOnly: true, Path: "/"}
	return newCookie
}
