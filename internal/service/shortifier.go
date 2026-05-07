package service

import (
	"context"

	flagvalues "go-url-shortener/internal/config/flag_values"
	"go-url-shortener/internal/repository"
	"go-url-shortener/internal/service/shortifier"
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
