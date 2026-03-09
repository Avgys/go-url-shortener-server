package service

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"net/url"
	"strings"
	"sync"
	"time"

	flagvalues "github.com/Avgys/go-url-shortener-server/internal/config/flag_values"
	"github.com/Avgys/go-url-shortener-server/internal/logger"
	"github.com/Avgys/go-url-shortener-server/internal/model"
	"github.com/Avgys/go-url-shortener-server/internal/repository"
	"github.com/Avgys/go-url-shortener-server/internal/shared"
	httpShared "github.com/Avgys/go-url-shortener-server/internal/shared/http"
	"github.com/rs/zerolog"
	"github.com/samber/lo"
	"golang.org/x/sync/errgroup"
)

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

func NewShortifier(done context.Context, stringGenerator StringGenerator, store repository.Repository, redirectAddr *flagvalues.NetAddress) *Shortifier {

	s := &Shortifier{done: done, stringGenerator: stringGenerator, store: store, redirectAddr: redirectAddr}

	g, c := errgroup.WithContext(done)

	s.initDeletePool(g, c)

	logger, closeLog := logger.NewLogger()
	s.logger = logger

	go func() {

		<-c.Done()

		if err := g.Wait(); err != nil {
			s.logger.Err(err).Send()
		}

		closeLog()
	}()

	return s
}

func (s *Shortifier) ResolveShortURL(ctx context.Context, inputURL string, traceLogger *zerolog.Logger) (*model.DBURL, error) {

	shortURL := strings.TrimSpace(inputURL)

	url, err := s.store.ResolveShortURL(ctx, shortURL)

	if err != nil {

		if errors.Is(err, repository.ErrNotFound) {
			traceLogger.Info().
				Str("repository error", err.Error()).
				Msg("url not found in repository")

			return url, httpShared.NewError("url not found", http.StatusNotFound)
		}

		traceLogger.Err(err).
			Send()

		return nil, err
	}

	if url.DeletedAtUTC != nil {
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

	const maxStoreRetryCount = 20

	for range maxStoreRetryCount {
		clear(full2shortBatch)

		const shortURLMaxLength = 8

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

	dbURLs = lo.Filter(dbURLs, func(h *model.DBURL, _ int) bool { return h.DeletedAtUTC == nil })

	urls := lo.Map(dbURLs, func(dbURL *model.DBURL, _ int) model.URLPair {
		link, _ := url.JoinPath(s.redirectAddr.String(), dbURL.ShortURL)
		return model.URLPair{ShortURL: link, OriginalURL: dbURL.OriginalURL}
	})

	return urls, err
}

func (s *Shortifier) startDelete(g *errgroup.Group, ctx context.Context, batchQueue chan *deleteMessage) {

	chanForDelete := make(chan []*deleteMessage)

	g.Go(func() error {
		defer close(chanForDelete)

		queueForDelete := make([]*deleteMessage, 0)

		const deleteTickerInterval = 1 * time.Second
		const minDeleteBatchSize = 200

		ticker := time.NewTicker(deleteTickerInterval)
		defer ticker.Stop()

		isLastDelete := false

		for {
			select {
			case <-ctx.Done():
				isLastDelete = true
			case <-ticker.C:
			case message, ok := <-batchQueue:
				if ok {
					queueForDelete = append(queueForDelete, message)
				} else {
					isLastDelete = true
				}

				if len(queueForDelete) < minDeleteBatchSize && !isLastDelete {
					continue
				}
			}

			chanForDelete <- queueForDelete

			if isLastDelete {
				return nil
			}

			queueForDelete = make([]*deleteMessage, 0)
		}
	})

	g.Go(func() error {
		for {
			select {
			case <-ctx.Done():
				return nil

			case batch := <-chanForDelete:

				groupedByUser := groupBatchByUser(batch)
				err := s.deleteBatchFromDB(groupedByUser)

				if err != nil {
					return err
				}
			}
		}
	})
}

func (s *Shortifier) deleteBatchFromDB(groupedByUser groupedBatch) error {

	if len(groupedByUser) == 0 {
		return nil
	}

	_, err := s.store.DeleteURLS(context.Background(), groupedByUser)
	return err
}

func groupBatchByUser(queueDelete []*deleteMessage) groupedBatch {
	if len(queueDelete) == 0 {
		return nil
	}

	groupedByUser := make(groupedBatch, 0)

	for _, message := range queueDelete {

		if message == nil {
			continue
		}

		urls, exists := groupedByUser[message.userID]

		if !exists {
			groupedByUser[message.userID] = message.shortURLs
		} else {
			groupedByUser[message.userID] = append(urls, message.shortURLs...)
		}
	}

	return groupedByUser
}

func (s *Shortifier) DeleteUrls(ctx context.Context, userID int64, urls []string, traceLogger *zerolog.Logger) error {

	select {
	case <-s.done.Done():
		return httpShared.NewError("server closing", http.StatusServiceUnavailable)
	// receiving channel from pool
	case queue := <-s.deletePool:
		defer func() { s.deletePool <- queue }()

		*queue <- &deleteMessage{userID: userID, shortURLs: urls}

		return nil
	default:
		traceLogger.Warn().
			Int("urls_count", len(urls)).
			Int64("user_id", userID).
			Msg("delete queue is full; cannot enqueue delete request")

		return httpShared.NewError("delete queue is full, try again later", http.StatusTooManyRequests)
	}
}

func (s *Shortifier) initDeletePool(g *errgroup.Group, ctx context.Context) {
	const maxDeleteWorkers = 10

	s.deletePool = make(chan *deleteQueue, maxDeleteWorkers)
	deleteChannels := make([]*deleteQueue, 0, maxDeleteWorkers)

	for range maxDeleteWorkers {
		deleteChannel := make(deleteQueue)

		s.deletePool <- &deleteChannel
		deleteChannels = append(deleteChannels, &deleteChannel)
	}

	go func() {
		<-ctx.Done()

		for _, c := range deleteChannels {
			close(*c)
		}
	}()

	batchDeleteCh := fanIn(ctx, deleteChannels)

	s.startDelete(g, ctx, batchDeleteCh)
}

func fanIn(doneCtx context.Context, resultChs []*deleteQueue) chan *deleteMessage {
	finalCh := make(chan *deleteMessage)

	var wg sync.WaitGroup

	for _, ch := range resultChs {
		chClosure := ch

		wg.Add(1)

		go func() {
			defer wg.Done()

			for data := range *chClosure {
				select {
				case <-doneCtx.Done():
					return
				case finalCh <- data:
				}
			}
		}()
	}

	go func() {
		wg.Wait()
		close(finalCh)
	}()

	return finalCh
}
