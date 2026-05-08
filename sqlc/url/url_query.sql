-- name: GetURLByShortURL :one
SELECT id, short_url, long_url, created_at, deleted_at_utc
FROM urls
WHERE short_url = $1
LIMIT 1;

-- name: ListURLsByUserID :many
SELECT id, short_url, long_url, created_at, user_id, deleted_at_utc
FROM urls
WHERE user_id = $1;
