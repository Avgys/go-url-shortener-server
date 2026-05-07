package shortifier

import (
	"context"
	"errors"

	flagvalues "github.com/Avgys/go-url-shortener-server/internal/config/flag_values"
	"github.com/Avgys/go-url-shortener-server/internal/model/requests"
	"github.com/Avgys/go-url-shortener-server/internal/repository"
	"github.com/rs/zerolog"
)

var (
	ErrCollision = errors.New("could not find free space to store url")
	ErrConflict  = errors.New("url already stored")
)

type ShortenBatchReq struct {
	UserID int64
	URLs   requests.ShortenBatchReq
}

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

type Shortifier struct {
	domain          string
	store           repository.Repository
	stringGenerator StringGenerator
	redirectAddr    *flagvalues.NetAddress

	done       context.Context
	deletePool chan *deleteQueue
	logger     *zerolog.Logger
}

type groupedBatch map[int64][]string
