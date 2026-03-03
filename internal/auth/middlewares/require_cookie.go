package middlewares

import (
	"context"
	"net/http"
	"time"

	"github.com/Avgys/go-url-shortener-server/internal/auth"
	"github.com/Avgys/go-url-shortener-server/internal/auth/jwttoken"
	"github.com/Avgys/go-url-shortener-server/internal/logger"
)

func RequireCookie(h http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {

		traceLogger := logger.Middleware(r.Context(), "RequireCookie")

		authCookie, err := r.Cookie(auth.AuthCookie)

		if err == http.ErrNoCookie || authCookie == nil || authCookie.Value == "" {

			traceLogger.Info().
				Str("Cookie", "No-Cookie").
				Send()

			w.WriteHeader(http.StatusNoContent)
			return
		}

		claims, err := jwttoken.ParseToken(authCookie.Value)

		if err != nil || claims == nil || claims.UserID == 0 {

			traceLogger.Info().
				Str("AuthCookie", "Empty claims").
				Send()

			w.WriteHeader(http.StatusUnauthorized)
			return
		}

		// refresh cookie expire time
		newCookie := &http.Cookie{Name: auth.AuthCookie, Value: authCookie.Value, Expires: time.Now().Add(jwttoken.TOKEN_EXP)}

		authCtx := context.WithValue(r.Context(), auth.Claims, claims)
		r = r.WithContext(authCtx)

		http.SetCookie(w, newCookie)

		h.ServeHTTP(w, r)

		traceLogger.Info().
			Str("AuthCookie", authCookie.Value).
			Send()
	})
}
