package audit

import (
	"context"
	"errors"
	"fmt"
	"go-url-shortener/internal/config"
	"time"

	"github.com/rs/zerolog"
)

type Observer interface {
	Update(event AuditEvent) error
	GetID() int
}

type AuditService struct {
	observers map[int]Observer
	logger    *zerolog.Logger
}

func (a *AuditService) Init(done context.Context, cfg *config.Config, traceLogger *zerolog.Logger) {
	if len(cfg.AuditFile) > 0 {
		auditFile := NewAuditFile(done, 1, cfg.AuditFile, traceLogger)

		if err := a.Register(auditFile); err != nil {
			traceLogger.Err(err).Msg("register audit file observer")
		}
	}

	if len(cfg.AuditURL) > 0 {
		auditClient := NewAuditClient(done, 2, cfg.AuditURL, traceLogger)

		if err := a.Register(auditClient); err != nil {
			traceLogger.Err(err).Msg("register audit client observer")
		}
	}
}

func NewAuditService(logger *zerolog.Logger, observers ...Observer) *AuditService {
	observersMap := make(map[int]Observer)
	for _, observer := range observers {
		observersMap[observer.GetID()] = observer
	}

	return &AuditService{logger: logger, observers: observersMap}
}

type AuditEvent struct {
	Ts     int64  `json:"ts" csv:"ts"`
	Action string `json:"action" csv:"action"`
	UserID string `json:"user_id" csv:"user_id"`
	URL    string `json:"url" csv:"url"`
}

func (a *AuditService) Register(o Observer) error {
	if o == nil {
		return fmt.Errorf("observer is nil")
	}

	_, exists := a.observers[o.GetID()]

	if exists {
		return fmt.Errorf("observer with id %v already exists", o.GetID())
	}

	a.observers[o.GetID()] = o

	return nil
}

func (a *AuditService) Unregister(o Observer) error {
	if o == nil {
		return fmt.Errorf("observer is nil")
	}

	delete(a.observers, o.GetID())

	return nil
}

func (a *AuditService) Publish(ctx context.Context, action string, userID string, url string) {
	event := AuditEvent{
		Ts:     time.Now().Unix(),
		Action: action,
		UserID: userID,
		URL:    url,
	}

	pubErrors := make([]error, 0)

	go func() {
		for _, observer := range a.observers {
			err := observer.Update(event)

			if err != nil {
				pubErrors = append(pubErrors, err)
			}
		}

		var returnErr error = nil

		if len(pubErrors) > 0 {
			returnErr = fmt.Errorf("Publish errors: %w", errors.Join(pubErrors...))
		}

		if returnErr != nil {
			a.logger.Err(returnErr).Send()
		}
	}()
}
