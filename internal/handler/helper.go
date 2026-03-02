package handler

import (
	"context"
	"errors"
	"net/http"

	"github.com/Avgys/go-url-shortener-server/internal/auth"
	"github.com/Avgys/go-url-shortener-server/internal/auth/jwt_token"
	"github.com/Avgys/go-url-shortener-server/internal/model"
	"github.com/Avgys/go-url-shortener-server/internal/service"
	shared "github.com/Avgys/go-url-shortener-server/internal/shared/http"
	"github.com/rs/zerolog"
)

func getBody(w http.ResponseWriter, r *http.Request, traceLogger *zerolog.Logger) ([]byte, bool) {
	body, err := shared.GetRequestBody(w, r)

	if err != nil {
		shared.WriteError(w, r, err, traceLogger)
		return nil, true
	}

	return body, false
}

func getShortURL(h *Handlers, url string, traceLogger *zerolog.Logger, w http.ResponseWriter, r *http.Request) (*model.IndexedShortURL, bool) {
	batch := model.ShortenBatchReq{model.IndexedFullURL{FullURL: url}}
	urls, shouldReturn := shortenBatch(h, batch, traceLogger, w, r)

	if shouldReturn {
		return nil, true
	}

	return &(*urls)[0], false
}

func shortenBatch(h *Handlers, model model.ShortenBatchReq, traceLogger *zerolog.Logger, w http.ResponseWriter, r *http.Request) (*model.ShortenBatchResp, bool) {

	ctx := r.Context()
	claims, _ := getClaims(ctx)

	serviceReq := &service.ShortenBatchReq{URLs: model, UserID: claims.UserID}
	resultURL, err := h.Shortifier.ShortifyBatch(r.Context(), serviceReq, traceLogger)

	if err != nil {
		if errors.Is(err, service.ErrCollision) {
			err = shared.NewError(err.Error(), http.StatusTooManyRequests)
		}

		shared.WriteError(w, r, err, traceLogger)
		return nil, true
	}

	return &resultURL, false
}

func getClaims(ctx context.Context) (*jwt_token.Claims, bool) {
	claims, ok := ctx.Value(auth.CLAIMS).(*jwt_token.Claims)

	if !ok || claims == nil || claims.UserID == 0 {
		return &jwt_token.Claims{}, false
	}

	return claims, true
}
