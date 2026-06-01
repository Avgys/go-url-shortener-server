package handler

import (
	"encoding/json"
	"net/http"
	"strconv"

	"go-url-shortener/internal/logger"
	"go-url-shortener/internal/service/auth"
	httphelper "go-url-shortener/internal/shared/http"

	"github.com/go-chi/chi/v5"
)

// Redirect handles GET /{shortURL} and responds with 307 Temporary Redirect to the original URL.
// When the request context carries auth claims, an audit "follow" event is published asynchronously.
func (h *Handlers) Redirect(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	traceLogger := logger.FromContext(ctx, logger.GetFuncName())

	url := chi.URLParam(r, "url")
	dbURL, err := h.Shortifier.ResolveShortURL(ctx, url, traceLogger)

	if httphelper.HandleErr(w, r, err, traceLogger) {
		return
	}

	claims, err := auth.GetFromContext(ctx)

	traceLogger.Err(err).Send()

	if err == nil {
		h.AuditService.Publish(ctx, "follow", strconv.FormatInt(claims.UserID, 10), dbURL.OriginalURL)
	}

	w.Header().Set("Location", dbURL.OriginalURL)
	httphelper.WriteResponse(w, nil, http.StatusTemporaryRedirect, traceLogger)
}

// GetURLsByUserID handles GET /api/user/urls and returns the authenticated user's URL pairs as JSON.
// It responds with 204 No Content when the user has no URLs.
func (h *Handlers) GetURLsByUserID(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	traceLogger := logger.FromContext(ctx, logger.GetFuncName())

	claims, err := auth.GetFromContext(ctx)

	if err != nil {
		traceLogger.Err(err).Send()
		err = httphelper.NewError("unauthorized/broken token", http.StatusUnauthorized)
		if httphelper.HandleErr(w, r, err, traceLogger) {
			return
		}
	}

	urls, err := h.Shortifier.GetURLsByUserID(ctx, claims.UserID, traceLogger)
	if err != nil {
		if httphelper.HandleErr(w, r, err, traceLogger) {
			return
		}
	}

	var status int
	var response []byte
	if len(urls) == 0 {
		status = http.StatusNoContent
	} else {
		response, _ = json.Marshal(urls)
		status = http.StatusOK
	}

	w.Header().Set("Content-type", "application/json")
	httphelper.WriteResponse(w, response, status, traceLogger)
}
