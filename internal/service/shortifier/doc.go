// Package shortifier implements URL shortening business logic.
//
// # Responsibilities
//
// A [Shortifier] maps long URLs to random short keys via [StringGenerator],
// persists mappings through [repository.Repository], and builds public short links
// using the configured redirect base URL ([flagvalues.NetAddress]).
// Successful shortens trigger audit events through [audit.Publisher].
//
// Deletions are asynchronous: [Shortifier.DeleteUrls] enqueues work on a pool of
// background workers that batch soft-deletes in the store.
//
// # HTTP integration
//
// Handlers in internal/handler typically call:
//
//   - [Shortifier.ShortenURL] — single URL; reads auth from *http.Request
//   - [Shortifier.ShortenBatch] — batch JSON; reads auth from *http.Request
//   - [Shortifier.ResolveShortURL] — redirect lookup by short key
//   - [Shortifier.GetURLsByUserID] — list URLs for a user
//   - [Shortifier.DeleteUrls] — queue deletion of short keys
//
// [Shortifier.ShortifyBatch] is the core store operation and accepts an explicit
// [context.Context] and [ShortenBatchReq] (including user ID).
//
// # Errors
//
// Package-level errors used by handlers to map HTTP status codes:
//
//   - [ErrCollision] — no free short key after retries (often 429)
//   - [ErrConflict] — long URL already stored for the user (409)
//
// [NewShortifier] also starts the delete worker pool; cancel the done context passed
// to NewShortifier to initiate graceful shutdown.
package shortifier
