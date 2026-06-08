package shortifier

import (
	"context"
	"errors"

	flagvalues "go-url-shortener/internal/config/flag_values"
	"go-url-shortener/internal/model/requests"
	"go-url-shortener/internal/repository"
	"go-url-shortener/internal/service/audit"

	"github.com/rs/zerolog"
)

var (
	// ErrCollision is returned when no free short key could be allocated after retries.
	ErrCollision = errors.New("could not find free space to store url")
	// ErrConflict is returned when the long URL is already stored for the user.
	ErrConflict = errors.New("url already stored")
)

// ShortenBatchReq is the service-layer batch shorten request with an owning user ID.
type ShortenBatchReq struct {
	UserID int64
	URLs   requests.ShortenBatchReq
}

// StringGenerator produces random short URL keys of a requested length.
//
//go:generate mockgen -destination=mocks/mock_string_generator.go -package=mocks go-url-shortener/internal/service/shortifier StringGenerator
type StringGenerator interface {
	GetRandomString(n int) string
}

type deleteMessage struct {
	shortURLs []string
	userID    int64
}

type deleteQueue chan *deleteMessage

type storeInfo struct {
	shortURL string
	new      bool
}

// Shortifier creates short links, resolves redirects, lists user URLs, and queues deletions.
//
// Create with [NewShortifier]. The value is safe for concurrent use by HTTP handlers
// once constructed; delete workers run until the constructor's done context is cancelled.
type Shortifier struct {
	store           repository.Repository
	stringGenerator StringGenerator
	auditService    audit.Publisher

	redirectAddr *flagvalues.NetAddress

	done       context.Context
	deletePool chan *deleteQueue
	logger     *zerolog.Logger
}

type groupedBatch map[int64][]string
