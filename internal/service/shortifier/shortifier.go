package shortifier

import (
	"context"

	flagvalues "go-url-shortener/internal/config/flag_values"
	"go-url-shortener/internal/logger"
	"go-url-shortener/internal/repository"

	"golang.org/x/sync/errgroup"
)

func NewShortifier(done context.Context, stringGenerator StringGenerator, store repository.Repository, redirectAddr *flagvalues.NetAddress) (*Shortifier, error) {

	s := &Shortifier{done: done, stringGenerator: stringGenerator, store: store, redirectAddr: redirectAddr}

	g, c := errgroup.WithContext(done)

	s.initDeletePool(g, c)

	log, closeLog, err := logger.NewBaseLogger(logger.GetFuncName())

	if err != nil {
		return nil, err
	}

	s.logger = log

	go func() {

		<-c.Done()

		if err := g.Wait(); err != nil {
			s.logger.Err(err).Send()
		}

		_ = closeLog()
	}()

	return s, nil
}
