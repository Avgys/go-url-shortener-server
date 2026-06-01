// Package shortifier implements URL shortening business logic.
//
// A Shortifier maps long URLs to random short keys, persists them through
// repository.Repository, builds public short links from the configured base URL,
// and publishes audit events when links are created. Deletions are processed
// asynchronously by a worker pool that batches store updates.
//
// Handlers in internal/handler call the request-aware helpers ShortenURL and
// ShortenBatch, which read the authenticated user from the request context.
// Lower-level methods such as ShortifyBatch and ResolveShortURL accept an
// explicit context and are used internally or for redirects and lookups.
package shortifier
