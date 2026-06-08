package audit

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"strconv"
	"sync"
	"sync/atomic"

	"github.com/rs/zerolog"
	"resty.dev/v3"
)

// ErrRetryAfter indicates the audit server responded with HTTP 429 Too Many Requests.
type ErrRetryAfter struct {
	error
	RetryAfter int64
}

// AuditClient POSTs [AuditEvent] JSON to a remote audit endpoint.
type AuditClient struct {
	httpClient *resty.Client
	logger     *zerolog.Logger

	isClosed  atomic.Bool
	closeOnce sync.Once
	auditHost string
}

// NewAuditClient creates an HTTP observer. It closes when ctx is cancelled.
func NewAuditClient(ctx context.Context, auditHost string, logger *zerolog.Logger) *AuditClient {

	newLogger := logger.With().
		Str("audit", "http_client").
		Str("audit_client_host", auditHost).
		Logger()

	audit := &AuditClient{httpClient: resty.New(), logger: &newLogger, auditHost: auditHost}

	go func() {
		<-ctx.Done()

		_ = audit.Close()
	}()

	return audit
}

// Update sends event to the configured audit host.
func (a *AuditClient) Update(ctx context.Context, event AuditEvent) error {
	return a.Send(ctx, event)
}

// Send POSTs auditEvent to auditHost and maps HTTP status codes to errors.
func (a *AuditClient) Send(ctx context.Context, auditEvent AuditEvent) error {
	if a == nil {
		return errors.New("audit service client is nil")
	}

	if a.isClosed.Load() {
		return errors.New("audit client closed")
	}

	if a.httpClient == nil {
		err := errors.New("audit service client is nil")
		a.logger.Err(err).Send()

		return err
	}

	if a.auditHost == "" {
		return fmt.Errorf("auditHost is empty")
	}

	a.logger.Debug().
		Msg("sending request to audit server")

	resp, err := a.httpClient.R().
		SetContext(ctx).
		SetBody(auditEvent).
		Post(a.auditHost)

	if err != nil {
		a.logger.Error().
			Err(err).
			Str("url", auditEvent.URL).
			Msg("audit request failed")

		return err
	}

	defer func() { _ = resp.Body.Close() }()

	a.logger.Debug().
		Int("status", resp.StatusCode()).
		Msg("audit server response")

	if resp.StatusCode() == http.StatusTooManyRequests {
		retryAfterSecs := resp.Header().Get("Retry-After")
		secs, _ := strconv.ParseInt(retryAfterSecs, 10, 32)
		a.logger.Warn().Int64("retry_after", secs).Msg("audit server rate limited")

		return ErrRetryAfter{error: errors.New("too many requests"), RetryAfter: secs}
	}

	if resp.StatusCode() == http.StatusOK {
		return nil
	}

	err = fmt.Errorf("unsupported error %s", resp.Status())
	a.logger.Error().Err(err).Msg("unexpected audit server response")

	return err
}

// Close releases the underlying HTTP client. Safe to call more than once.
func (a *AuditClient) Close() (err error) {
	a.closeOnce.Do(func() {
		a.isClosed.Store(true)

		if a.httpClient != nil {
			err = a.httpClient.Close()
		}
	})

	return err
}
