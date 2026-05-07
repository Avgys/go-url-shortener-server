package middlewares

import (
	"fmt"
	"net/http"

	"avgys-gophermat/internal/logger"
	"avgys-gophermat/internal/service/auth"
)

func RequireCookie(h http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		traceLogger, close, err := logger.Middleware(r.Context(), "compress")

		if err != nil {
			fmt.Print("couldn't create request logger")

			http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
			return
		}

		defer func() { _ = close() }()

		authCookie, err := r.Cookie(string(auth.CookieName))

		if err == http.ErrNoCookie || authCookie == nil || authCookie.Value == "" {

			traceLogger.Info().
				Str("Cookie", "No-Cookie").
				Send()

			w.WriteHeader(http.StatusUnauthorized)
			return
		}

		claims, err := auth.ParseToken(authCookie.Value)

		if err != nil || claims == nil || claims.UserID == 0 {

			traceLogger.Info().
				Str("AuthCookie", "Empty claims").
				Send()

			w.WriteHeader(http.StatusUnauthorized)
			return
		}

		if err := claims.InjectCookie(w); err != nil {
			traceLogger.Err(err).Msg("inject auth cookie")
			w.WriteHeader(http.StatusInternalServerError)
			return
		}

		authCtx := claims.WithContext(r.Context())
		r = r.WithContext(authCtx)

		h.ServeHTTP(w, r)
	})
}
