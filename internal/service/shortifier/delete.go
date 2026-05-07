package shortifier

import (
	"context"
	"net/http"
	"sync"
	"time"

	"github.com/rs/zerolog"
	httpShared "go-url-shortener/internal/shared/http"
	"golang.org/x/sync/errgroup"
)

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
		chClosure := *ch

		wg.Add(1)

		go func() {
			defer wg.Done()

			for {
				select {
				case <-doneCtx.Done():
					return
				case data, open := <-chClosure:
					if open {
						finalCh <- data
					} else {
						return
					}
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
