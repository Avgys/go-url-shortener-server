package middlewares

import (
	"fmt"
	"net/http"

	"go-url-shortener/internal/logger"
	"go-url-shortener/internal/service/auth"
)

func AuthRequireCookie(isRequired bool) func(h http.Handler) http.Handler {
	return func(h http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			traceLogger, close, err := logger.Middleware(r.Context(), "compress")

			if err != nil {
				fmt.Print("couldn't create request logger")

				http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
				return
			}

			defer func() { _ = close() }()

			authCookie, err := r.Cookie(string(auth.CookieName))

			if err != nil {
				traceLogger.Info().
					Str("AuthCookie", "Empty cookie").
					Send()
			}

			var claims *auth.TokenClaims

			if err == nil && authCookie != nil && authCookie.Value != "" {
				claims, err = auth.ParseToken(authCookie.Value)

				if err != nil {
					traceLogger.Info().
						Str("AuthCookie", "Empty claims").
						Send()
				}
			}

			if claims == nil && isRequired {
				w.WriteHeader(http.StatusUnauthorized)
				return
			}

			if claims != nil {
				if err := claims.InjectCookie(w); err != nil {
					traceLogger.Err(err).Msg("inject auth cookie")
					w.WriteHeader(http.StatusInternalServerError)
					return
				}

				authCtx := claims.WithContext(r.Context())
				r = r.WithContext(authCtx)
			}

			h.ServeHTTP(w, r)
		})
	}
}
