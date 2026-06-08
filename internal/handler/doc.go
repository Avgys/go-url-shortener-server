// Package handler implements the HTTP API for the URL shortener service.
//
// Handlers translate net/http requests into calls on the shortifier, storage,
// and audit layers. They write responses through internal/shared/http helpers
// and use the request-scoped logger from internal/logger.
//
// Routes are registered in internal/router. Typical mapping:
//
//   - POST /                      — ShortifyURL (plain-text body)
//   - POST /api/shorten            — ShortenURL (JSON)
//   - POST /api/shorten/batch      — ShortenBatch (JSON)
//   - GET  /{shortURL}             — Redirect
//   - GET  /api/user/urls          — GetURLsByUserID
//   - DELETE /api/user/urls        — DeleteShortURL (JSON array of short keys)
//   - GET  /ping                   — Ping
//
// Endpoints that create or list user URLs expect a valid JWT in the auth cookie;
// middleware in internal/middlewares attaches claims to the request context before
// the handler runs. Redirect may record an audit event when claims are present.
package handler
