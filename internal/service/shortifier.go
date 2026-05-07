package service

import (
	"context"

	flagvalues "github.com/Avgys/go-url-shortener-server/internal/config/flag_values"
	"github.com/Avgys/go-url-shortener-server/internal/repository"
	"github.com/Avgys/go-url-shortener-server/internal/service/shortifier"
)

type (
	Shortifier      = shortifier.Shortifier
	ShortenBatchReq = shortifier.ShortenBatchReq
)

var (
	ErrCollision = shortifier.ErrCollision
	ErrConflict  = shortifier.ErrConflict
)

func NewShortifier(done context.Context, stringGenerator shortifier.StringGenerator, store repository.Repository, redirectAddr *flagvalues.NetAddress) *Shortifier {
	return shortifier.NewShortifier(done, stringGenerator, store, redirectAddr)
}
