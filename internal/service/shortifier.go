package service

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"net/url"
	"strings"
	"time"

	flagvalues "github.com/Avgys/go-url-shortener-server/internal/config/flag_values"
	"github.com/Avgys/go-url-shortener-server/internal/model"
	"github.com/Avgys/go-url-shortener-server/internal/repository"
	"github.com/Avgys/go-url-shortener-server/internal/shared"
	httpShared "github.com/Avgys/go-url-shortener-server/internal/shared/http"
	"github.com/rs/zerolog"
	"github.com/samber/lo"
)

const shortURLMaxLength = 8
const maxStoreRetryCount = 20

var (
	ErrCollision = errors.New("could not find free space to store url")
	ErrConflict  = errors.New("url already stored")
)

type ShortenBatchReq struct {
	UserID int64
	URLs   []model.IndexedFullURL
}

type StringGenerator interface {
	GetRandomString(n int) string
}

type deleteMessage struct {
	shortURLs []string
	userID    int64
}

type Shortifier struct {
	domain          string
	store           repository.Repository
	stringGenerator StringGenerator
	redirectAddr    *flagvalues.NetAddress

	done        context.Context
	deleteQueue chan *deleteMessage
}

type storeInfo struct {
	shortURL string
	new      bool
}

func NewShortifier(done context.Context, stringGenerator StringGenerator, store repository.Repository, redirectAddr *flagvalues.NetAddress) *Shortifier {

	s := &Shortifier{done: done, stringGenerator: stringGenerator, store: store, redirectAddr: redirectAddr}

	s.deleteQueue = make(chan *deleteMessage, 50)
	s.startDeleteCoroutine()

	go func() {
		<-done.Done()
		close(s.deleteQueue)
	}()

	return s
}

func (s *Shortifier) ResolveShortURL(ctx context.Context, inputURL string, traceLogger *zerolog.Logger) (*model.DBURL, error) {

	shortURL := strings.TrimSpace(inputURL)

	url, err := s.store.ResolveShortURL(ctx, shortURL)

	if err != nil && errors.Is(err, repository.ErrNotFound) {
		traceLogger.Info().
			Str("repository error", err.Error()).
			Msg("url not found in repository")

		return url, httpShared.NewError("url not found", http.StatusNotFound)
	}

	if url.DeleteAtUTC != nil {
		return nil, httpShared.NewError("link deleted", http.StatusGone)
	}

	return url, err
}

func (s *Shortifier) ShortifyBatch(ctx context.Context, req *ShortenBatchReq, traceLogger *zerolog.Logger) (result model.ShortenBatchResp, err error) {

	urls := req.URLs

	if len(urls) == 0 {
		return nil, httpShared.NewError("empty url", http.StatusOK)
	}

	for _, url := range urls {

		if _, err := shared.GetURL(url.FullURL, true); err != nil {
			return nil, httpShared.NewError("url in wrong format", http.StatusBadRequest)
		}

		url.FullURL = strings.TrimSpace(url.FullURL)
	}

	readyBatch := make(map[string]storeInfo, len(urls))
	full2shortBatch := make(map[string]string, len(urls))
	urlsToInsert := lo.Map(urls, func(x model.IndexedFullURL, _ int) string { return x.FullURL })

	for range maxStoreRetryCount {
		clear(full2shortBatch)

		for _, fullURL := range urlsToInsert {
			full2shortBatch[fullURL] = s.stringGenerator.GetRandomString(shortURLMaxLength)
		}

		retryToInsert, alreadyExists, storeErr := s.store.StoreBatch(ctx, full2shortBatch, req.UserID)

		traceLogger.Info().
			Strs("retryToInsert", retryToInsert).
			Strs("alreadyExists", lo.Keys(alreadyExists)).
			Send()

		if storeErr != nil {
			return nil, fmt.Errorf("error saving short url in store, %w", storeErr)
		}

		// if already added to result or need to retry, then don't add to result
		for k, v := range full2shortBatch {
			_, ready := readyBatch[k]
			if ready || lo.Contains(retryToInsert, k) {
				continue
			}

			if shortURL, storedOld := alreadyExists[k]; storedOld {
				readyBatch[k] = storeInfo{shortURL: shortURL, new: false}
			} else {
				readyBatch[k] = storeInfo{shortURL: v, new: true}
			}
		}

		// If no need retry, break loop
		if len(retryToInsert) == 0 {
			urlsToInsert = nil
			break
		}

		urlsToInsert = retryToInsert
	}

	result = lo.Map(urls, func(x model.IndexedFullURL, _ int) model.IndexedShortURL {
		storedInfo := readyBatch[x.FullURL]
		link, _ := url.JoinPath(s.redirectAddr.String(), storedInfo.shortURL)
		return model.IndexedShortURL{CorrelationID: x.CorrelationID, ShortURL: link, IsCreated: storedInfo.new}
	})

	if len(urlsToInsert) != 0 {
		err = fmt.Errorf("no free space in storage, %w", err)
	}

	return
}

func (s *Shortifier) GetURLsByUserID(ctx context.Context, userID int64, traceLogger *zerolog.Logger) ([]model.URLPair, error) {

	dbURLs, err := s.store.GetURLsByUserID(ctx, userID)

	if err != nil && errors.Is(err, repository.ErrNotFound) {
		traceLogger.Info().
			Str("repository error", err.Error()).
			Msg("urls not found in repository")

		return nil, httpShared.NewError("no urls", http.StatusNoContent)
	}

	urls := lo.Map(dbURLs, func(dbURL *model.DBURL, _ int) model.URLPair {
		link, _ := url.JoinPath(s.redirectAddr.String(), dbURL.ShortURL)
		return model.URLPair{ShortURL: link, OriginalURL: dbURL.OriginalURL}
	})

	return urls, err
}

func (s *Shortifier) startDeleteCoroutine() {

	go func() {

		queueForDelete := make([]*deleteMessage, 50)
		ticker := time.NewTicker(5 * time.Second)

		for {
			select {
			case <-s.done.Done():
				s.deleteBatchFromDB(queueForDelete)
				clear(queueForDelete)
			case <-ticker.C:
				s.deleteBatchFromDB(queueForDelete)
				clear(queueForDelete)
			case message := <-s.deleteQueue:
				queueForDelete = append(queueForDelete, message)

				if len(queueForDelete) < 50 {
					continue
				}

				s.deleteBatchFromDB(queueForDelete)
				clear(queueForDelete)
			}
		}
	}()

}

func (s *Shortifier) deleteBatchFromDB(queueDelete []*deleteMessage) error {

	t := lo.Filter(queueDelete, func(m *deleteMessage, _ int) bool {
		return m != nil
	})

	if len(t) == 0 {
		return nil
	}

	groupedByUser := lo.SliceToMap(t, func(m *deleteMessage) (int64, []string) { return m.userID, m.shortURLs })
	return s.store.DeleteURLS(context.Background(), groupedByUser)
}

func (s *Shortifier) DeleteUrls(ctx context.Context, userID int64, urls []string, traceLogger *zerolog.Logger) error {
	go func() {
		s.deleteQueue <- &deleteMessage{userID: userID, shortURLs: urls}
	}()

	return nil
}
