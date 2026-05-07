package handler

import (
	"encoding/json"
	"errors"
	"fmt"
	"net/http"

	"github.com/Avgys/go-url-shortener-server/internal/auth/jwttoken"
	"github.com/Avgys/go-url-shortener-server/internal/model/requests"
	"github.com/Avgys/go-url-shortener-server/internal/model/responses"
	"github.com/Avgys/go-url-shortener-server/internal/service"
	shared "github.com/Avgys/go-url-shortener-server/internal/shared/http"
	"github.com/rs/zerolog"
)

func getBody(w http.ResponseWriter, r *http.Request) ([]byte, error) {
	body, err := shared.GetRequestBody(w, r)

	if err != nil {
		return nil, err
	}

	return body, nil
}

func getJSONBody(r *http.Request, value any) error {

	dec := json.NewDecoder(r.Body)
	err := dec.Decode(value)

	if err != nil {
		var syntaxErr *json.SyntaxError

		if errors.As(err, &syntaxErr) {
			err = shared.NewError(err.Error(), http.StatusBadRequest)
		}
	}

	return err
}

func getShortURL(h *Handlers, url string, traceLogger *zerolog.Logger, r *http.Request) (*responses.IndexedShortURL, error) {
	batch := requests.ShortenBatchReq{requests.IndexedFullURL{FullURL: url}}
	urls, err := shortenBatch(h, batch, traceLogger, r)

	if err != nil {
		return nil, err
	}

	return &(*urls)[0], nil
}

func shortenBatch(h *Handlers, batch requests.ShortenBatchReq, traceLogger *zerolog.Logger, r *http.Request) (*responses.ShortenBatchResp, error) {

	ctx := r.Context()
	claims, err := jwttoken.GetClaims(ctx)

	if err != nil {
		showErr := shared.NewError("unauthorized/broken token", http.StatusUnauthorized)
		return nil, fmt.Errorf("%w inner error: %w", showErr, err)
	}

	serviceReq := &service.ShortenBatchReq{URLs: batch, UserID: claims.UserID}
	resultURL, err := h.Shortifier.ShortifyBatch(r.Context(), serviceReq, traceLogger)

	if err != nil {
		if errors.Is(err, service.ErrCollision) {
			err = shared.NewError(err.Error(), http.StatusTooManyRequests)
		}

		return nil, err
	}

	return &resultURL, nil
}
