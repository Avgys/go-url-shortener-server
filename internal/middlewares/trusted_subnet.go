package middlewares

import (
	"fmt"
	"net"
	"net/http"
	"net/netip"

	"go-url-shortener/internal/handler"
	"go-url-shortener/internal/logger"
)

func WithTrustedSubnet(h *handler.Handlers) func(http.Handler) http.Handler {

	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {

			traceLogger, close, err := logger.Middleware(r.Context(), "trusted_subnet")

			if err != nil {
				fmt.Print("couldn't create request logger")

				http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
				return
			}

			defer func() { _ = close() }()

			trustedSubnet, err := netip.ParsePrefix(h.Config.TrustedSubnet)
			if err != nil {
				http.Error(w, http.StatusText(http.StatusBadRequest), http.StatusBadRequest)
				return
			}

			realIP, err := parseRemoteAddr(r.RemoteAddr)
			if err != nil {
				http.Error(w, http.StatusText(http.StatusBadRequest), http.StatusBadRequest)
				return
			}

			var statusCode = http.StatusContinue

			if trustedSubnet.Contains(realIP) {
				next.ServeHTTP(w, r)
			} else {
				statusCode = http.StatusForbidden
				http.Error(w, http.StatusText(http.StatusForbidden), http.StatusForbidden)
			}

			traceLogger.Info().
				Str("realip", realIP.String()).
				Str("trustedSubnet", trustedSubnet.String()).
				Str("status", http.StatusText(statusCode)).
				Send()
		})
	}
}

func parseRemoteAddr(remoteAddr string) (netip.Addr, error) {
	if addrPort, err := netip.ParseAddrPort(remoteAddr); err == nil {
		return addrPort.Addr(), nil
	}

	host, _, err := net.SplitHostPort(remoteAddr)
	if err != nil {
		host = remoteAddr
	}

	return netip.ParseAddr(host)
}
