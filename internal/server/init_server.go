package server

import (
	"context"
	"crypto/tls"
	"errors"
	"fmt"
	"io"
	"net"
	"net/http"
	"os"
	"strings"

	"go-url-shortener/internal/config"
	"go-url-shortener/internal/handler"
	"go-url-shortener/internal/repository"
	"go-url-shortener/internal/router"
	"go-url-shortener/internal/service"
	"go-url-shortener/internal/service/audit"
	"go-url-shortener/internal/service/shortifier"

	"github.com/rs/zerolog"
	"golang.org/x/crypto/acme/autocert"
)

type Server struct {
	Start    func() error
	Shutdown func(context.Context) error
}

func NewServer(done context.Context, traceLogger *zerolog.Logger) (*Server, error) {

	cfg, err := config.GetConfig(os.Args[1:], traceLogger)

	if err != nil {
		return nil, err
	}

	h, err := prepareDI(done, cfg, traceLogger)

	if err != nil {
		return nil, err
	}

	r := router.NewRouter(h)

	hosts := autocertHosts(cfg)

	tlsConfig, certManager, err := getTlsServer(cfg, hosts, traceLogger)
	if err != nil {
		return nil, err
	}

	srv := &http.Server{
		Addr:      cfg.AppURL.Host,
		Handler:   r,
		TLSConfig: tlsConfig,
	}

	var certSrv *http.Server
	if certManager != nil {
		certSrv = &http.Server{
			Addr:    ":http",
			Handler: certManager.HTTPHandler(srv.Handler),
		}
	}

	startSrv := getStartFn(cfg, certManager, certSrv, traceLogger, srv)
	shutdownFn := getShutdownFn(certManager, certSrv, srv)

	return &Server{
		Start:    startSrv,
		Shutdown: shutdownFn,
	}, nil
}

func getTlsServer(cfg *config.Config, hosts []string, traceLogger *zerolog.Logger) (*tls.Config, *autocert.Manager, error) {

	var tlsConfig *tls.Config
	var certManager *autocert.Manager

	if cfg.HttpsEnabled {
		if useSelfSignedTLS(hosts) {
			var err error
			tlsConfig, err = generateSelfSignedTLSConfig(hosts)

			if err != nil {
				return nil, nil, fmt.Errorf("self-signed TLS: %w", err)
			}

			traceLogger.Warn().
				Strs("hosts", hosts).
				Msg("using self-signed TLS; skipping autocert HTTP on :80")
		} else {
			certManager = newAutocertManager(hosts)
			tlsConfig = certManager.TLSConfig()
		}
	}

	return tlsConfig, certManager, nil
}

func getStartFn(cfg *config.Config, certManager *autocert.Manager, certSrv *http.Server, traceLogger *zerolog.Logger, srv *http.Server) func() error {
	var startSrv func() error
	if cfg.HttpsEnabled {
		startSrv = func() error {
			if certManager != nil {
				go func() {
					if err := certSrv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
						traceLogger.Error().Err(err).Msg("autocert HTTP server failed")
					}
				}()
			}

			return srv.ListenAndServeTLS("", "")
		}
	} else {
		startSrv = srv.ListenAndServe
	}

	return startSrv
}

func getShutdownFn(certManager *autocert.Manager, certSrv *http.Server, srv *http.Server) func(ctx context.Context) error {
	shutdownFn := func(ctx context.Context) error {
		var shutdownErrs error = nil
		if certManager != nil {
			if err := certSrv.Shutdown(ctx); err != nil {
				shutdownErrs = errors.Join(shutdownErrs, err)
			}
		}

		if err := srv.Shutdown(ctx); err != nil {
			shutdownErrs = errors.Join(shutdownErrs, err)
		}

		return shutdownErrs
	}

	return shutdownFn
}

func newAutocertManager(hosts []string) *autocert.Manager {
	return &autocert.Manager{
		Cache:      autocert.DirCache("cache-dir"),
		Prompt:     autocert.AcceptTOS,
		HostPolicy: autocert.HostWhitelist(hosts...),
	}
}

func autocertHosts(cfg *config.Config) []string {
	seen := make(map[string]struct{})
	hosts := make([]string, 0, 2)

	for _, hostPort := range []string{cfg.AppURL.Host, cfg.RedirectDomain.Host} {
		host := hostnameFromAddr(hostPort)
		if host == "" {
			continue
		}

		if _, ok := seen[host]; ok {
			continue
		}

		seen[host] = struct{}{}
		hosts = append(hosts, host)
	}

	return hosts
}

func hostnameFromAddr(hostPort string) string {
	host, _, err := net.SplitHostPort(hostPort)
	if err != nil {
		return strings.Trim(hostPort, ":")
	}

	return host
}

func prepareDI(done context.Context, cfg *config.Config, traceLogger *zerolog.Logger) (*handler.Handlers, error) {

	closers := make([]io.Closer, 0)

	go func() {
		<-done.Done()

		for i := len(closers) - 1; i >= 0; i-- {
			if err := closers[i].Close(); err != nil {
				traceLogger.Err(err).Str("message", "error when closing services")
			}
		}
	}()

	// Repos
	store, err := repository.NewRepository(done, cfg, traceLogger)

	if err != nil {
		return nil, fmt.Errorf("error initializing repository: %w", err)
	}

	closers = append(closers, store)

	//	Audit services initialization
	auditService := audit.NewAuditService(done, traceLogger)
	auditService.Init(cfg, traceLogger)
	closers = append(closers, auditService)

	generator := service.NewStringGenerator()
	shortifierService, err := shortifier.NewShortifier(done, generator, store, auditService, &cfg.RedirectDomain)

	if err != nil {
		return nil, fmt.Errorf("error initializing shortifier: %w", err)
	}

	h := handler.NewHandlers(shortifierService, store, auditService, cfg)

	return h, nil
}
