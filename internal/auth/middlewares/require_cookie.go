package middlewares

import (
	"net/http"

	"github.com/Avgys/go-url-shortener-server/internal/auth/jwttoken"
	"github.com/Avgys/go-url-shortener-server/internal/logger"
)

const authCookieName string = "AUTH_COOKIE"

func RequireCookie(h http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {

		traceLogger := logger.Middleware(r.Context(), "RequireCookie")

		authCookie, err := r.Cookie(string(authCookieName))

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
		newCookie := createCookie(authCookie.Value)

		authCtx := claims.WithContext(r.Context())
		r = r.WithContext(authCtx)

		http.SetCookie(w, newCookie)

		h.ServeHTTP(w, r)
	})
}
