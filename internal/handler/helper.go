package handler

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"

	"github.com/Avgys/go-url-shortener-server/internal/auth"
	"github.com/Avgys/go-url-shortener-server/internal/auth/jwttoken"
	"github.com/Avgys/go-url-shortener-server/internal/model"
	"github.com/Avgys/go-url-shortener-server/internal/service"
	shared "github.com/Avgys/go-url-shortener-server/internal/shared/http"
	"github.com/rs/zerolog"
)

func getBody(w http.ResponseWriter, r *http.Request, traceLogger *zerolog.Logger) ([]byte, error) {
	body, err := shared.GetRequestBody(w, r)

	if err != nil {
		return nil, err
	}

	return body, nil
}

func getJsonBody(r *http.Request, value any, traceLogger *zerolog.Logger) error {

	dec := json.NewDecoder(r.Body)
	err := dec.Decode(value)

	var syntaxErr json.SyntaxError

	if errors.As(err, &syntaxErr) {
		err = shared.NewError(err.Error(), http.StatusBadRequest)
	}

	return err
}

func getShortURL(h *Handlers, url string, traceLogger *zerolog.Logger, r *http.Request) (*model.IndexedShortURL, error) {
	batch := model.ShortenBatchReq{model.IndexedFullURL{FullURL: url}}
	urls, err := shortenBatch(h, batch, traceLogger, r)

	if err != nil {
		return nil, err
	}

	return &(*urls)[0], nil
}

func shortenBatch(h *Handlers, model model.ShortenBatchReq, traceLogger *zerolog.Logger, r *http.Request) (*model.ShortenBatchResp, error) {

	ctx := r.Context()
	claims, _ := getClaims(ctx)

	serviceReq := &service.ShortenBatchReq{URLs: model, UserID: claims.UserID}
	resultURL, err := h.Shortifier.ShortifyBatch(r.Context(), serviceReq, traceLogger)

	if err != nil {
		if errors.Is(err, service.ErrCollision) {
			err = shared.NewError(err.Error(), http.StatusTooManyRequests)
		}

		return nil, err
	}

	return &resultURL, nil
}

func getClaims(ctx context.Context) (*jwttoken.Claims, error) {
	claims, ok := ctx.Value(auth.Claims).(*jwttoken.Claims)

	if !ok || claims == nil || claims.UserID == 0 {
		return &jwttoken.Claims{}, errors.New("wrong auth token")
	}

	return claims, nil
}
