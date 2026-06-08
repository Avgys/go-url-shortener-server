package middlewares

import (
	"crypto/rand"
	"math/big"
	"net/http"
	"time"

	"go-url-shortener/internal/logger"
	"go-url-shortener/internal/service/auth"
)

func SetCookie(h http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		traceLogger, close, err := logger.Middleware(r.Context(), "SetCookie")
		if err != nil {
			http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
			return
		}
		defer func() { _ = close() }()

		authCookie, err := r.Cookie(auth.CookieName)

		resetCookie := false
		var claims *auth.TokenClaims

		if err == http.ErrNoCookie {
			resetCookie = true
		} else {
			claims, err = auth.ParseToken(authCookie.Value)

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

			claims = auth.NewToken(userID.Int64(), "")
			tokenString, err = claims.ToString()
			if err != nil || tokenString == "" {
				if err != nil {
					traceLogger.Error().Err(err).Msg("failed to create JWT token for auth cookie")
				} else {
					traceLogger.Error().Msg("generated invalid JWT token for auth cookie")
				}
				http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
				return
			}

			traceLogger.Info().Bool("IsNewCookie", true).Send()
		} else {
			tokenString = authCookie.Value
			traceLogger.Info().Bool("IsNewCookie", false).Send()
		}

		authCtx := claims.WithContext(r.Context())
		r = r.WithContext(authCtx)

		http.SetCookie(w, &http.Cookie{
			Name:     auth.CookieName,
			Value:    tokenString,
			Expires:  time.Now().Add(auth.TokenExp),
			HttpOnly: true,
			Path:     "/",
		})

		h.ServeHTTP(w, r)
	})
}
