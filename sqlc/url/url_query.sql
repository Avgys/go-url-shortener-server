-- name: GetURLByShortURL :one
SELECT id, short_url, long_url, created_at, user_id, deleted_at_utc
FROM urls
WHERE short_url = $1
LIMIT 1;

-- name: GetURLsByUserID :many
SELECT id, short_url, long_url, created_at, user_id, deleted_at_utc
FROM urls
WHERE user_id = $1;

-- name: StoreBatchUrls :many
SELECT
  CAST(f.short_url AS text) AS short_url,
  CAST(f.long_url AS text) AS long_url,
  CAST(f.stored_short_url AS text) AS stored_short_url,
  CAST(f.retry AS boolean) AS retry
FROM LATERAL public.store_batch_urls(sqlc.arg(long_urls)::text[], sqlc.arg(short_urls)::text[], sqlc.arg(user_id)) AS f(short_url, long_url, stored_short_url, retry);

-- name: GetURLsStats :one
SELECT
  COUNT(DISTINCT long_url)::bigint AS unique_long_url_count,
  COUNT(DISTINCT user_id)::bigint AS unique_user_id_count
FROM urls;
