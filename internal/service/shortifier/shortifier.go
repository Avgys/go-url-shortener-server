package shortifier

import (
	"context"

	flagvalues "github.com/Avgys/go-url-shortener-server/internal/config/flag_values"
	"github.com/Avgys/go-url-shortener-server/internal/logger"
	"github.com/Avgys/go-url-shortener-server/internal/repository"
	"golang.org/x/sync/errgroup"
)

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
