-- name: CreateUser :one
INSERT INTO public.users(login, hash_salt, password_hash)
VALUES ($1, $2, $3)
RETURNING Id;

-- name: GetUserByLogin :one
SELECT id, login, hash_salt, password_hash, created_at 
FROM public.users
where login = $1;