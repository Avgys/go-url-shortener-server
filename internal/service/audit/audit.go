package audit

import (
	"context"
	"fmt"
	"go-url-shortener/internal/config"
	"go-url-shortener/internal/service/fanout"
	"sync"
	"time"

	"github.com/rs/zerolog"
)

const singleObserverQueueLength = 10

// Observer receives audit events. Implementations include [AuditFile] and [AuditClient].
type Observer interface {
	Update(ctx context.Context, event AuditEvent) error
}

// AuditService fans out [AuditEvent] values to registered observers.
type AuditService struct {
	done      context.Context
	observers map[Observer]chan AuditEvent
	logger    *zerolog.Logger
	Fanout    *fanout.Fanout[AuditEvent]

	registerMux sync.Mutex
	wg          sync.WaitGroup
}

// AuditEvent is the payload published to observers and sent to remote audit APIs.
type AuditEvent struct {
	Ts     int64  `json:"ts" csv:"ts"`
	Action string `json:"action" csv:"action"`
	UserID string `json:"user_id" csv:"user_id"`
	URL    string `json:"url" csv:"url"`
}

// Init registers file and HTTP observers when cfg.AuditFile or cfg.AuditURL are set.
func (a *AuditService) Init(cfg *config.Config, traceLogger *zerolog.Logger) {
	if len(cfg.AuditFile) > 0 {
		auditFile := NewAuditFile(a.done, cfg.AuditFile, traceLogger)

		if err := a.Register(auditFile); err != nil {
			traceLogger.Err(err).Msg("register audit file observer")
		}
	}

	if len(cfg.AuditURL) > 0 {
		auditClient := NewAuditClient(a.done, cfg.AuditURL, traceLogger)

		if err := a.Register(auditClient); err != nil {
			traceLogger.Err(err).Msg("register audit client observer")
		}
	}
}

// NewAuditService creates an audit hub with optional pre-registered observers.
func NewAuditService(done context.Context, logger *zerolog.Logger, observers ...Observer) *AuditService {
	a := &AuditService{
		done:      done,
		logger:    logger,
		observers: make(map[Observer]chan AuditEvent),
		Fanout:    fanout.New[AuditEvent](done, 10, logger),
	}

	for _, observer := range observers {
		if err := a.Register(observer); err != nil {
			logger.Err(err).Msg("register audit observer")
		}
	}

	return a
}

// Register adds an observer. Returns an error if o is nil or already registered.
func (a *AuditService) Register(o Observer) error {
	a.registerMux.Lock()
	defer a.registerMux.Unlock()

	if o == nil {
		return fmt.Errorf("observer is nil")
	}

	if _, exists := a.observers[o]; exists {
		return fmt.Errorf("observer already exists: %T", o)
	}

	observerCh := a.Fanout.Take(fmt.Sprintf("%T", o), singleObserverQueueLength)
	a.observers[o] = observerCh

	a.wg.Add(1)
	go a.runObserver(a.done, o, observerCh)

	return nil
}

// Unregister removes an observer.
func (a *AuditService) Unregister(o Observer) error {
	a.registerMux.Lock()
	defer a.registerMux.Unlock()

	if o == nil {
		return fmt.Errorf("observer is nil")
	}

	ch, ok := a.observers[o]
	if !ok {
		return fmt.Errorf("observer not found: %T", o)
	}

	delete(a.observers, o)

	return a.Fanout.Close(ch)
}

// Publish enqueues an [AuditEvent] for observers. Drops the event when the fanout buffer is full.
func (a *AuditService) Publish(action string, userID string, url string) {
	event := AuditEvent{
		Ts:     time.Now().Unix(),
		Action: action,
		UserID: userID,
		URL:    url,
	}

	select {
	case a.Fanout.In <- event:
	default:
		a.logger.Warn().
			Int("queue_len", len(a.Fanout.In)).
			Int("queue_cap", cap(a.Fanout.In)).
			Msg("audit fanout input queue is full, event dropped")
	}
}

func (a *AuditService) runObserver(ctx context.Context, o Observer, ch chan AuditEvent) {
	defer a.wg.Done()

	for {
		select {
		case <-ctx.Done():
			return
		case event, ok := <-ch:
			if !ok {
				return
			}

			if err := o.Update(ctx, event); err != nil {
				a.logger.Err(err).Msg("audit observer update failed")
			}
		}
	}
}

// Close unregisters all observers and waits for observer goroutines to stop.
func (a *AuditService) Close() error {
	a.registerMux.Lock()

	for o, ch := range a.observers {
		delete(a.observers, o)

		if err := a.Fanout.Close(ch); err != nil {
			a.logger.Err(err).Msg("close audit observer channel")
		}
	}

	a.registerMux.Unlock()
	a.wg.Wait()

	return nil
}
