-- name: GetUnproccessedOrders :many
WITH picked AS (
	SELECT order_num
	FROM orders
	WHERE status IN (0, 1)
	ORDER BY status DESC, created_at
	LIMIT $1
	FOR UPDATE SKIP LOCKED
)
UPDATE orders o
SET status = 1
FROM picked p
WHERE o.order_num = p.order_num
RETURNING o.order_num, o.status, o.accrual, o.user_id, o.created_at, o.updated_at;

-- name: GetOrdersByUser :many
SELECT order_num, status, accrual, user_id, created_at, updated_at
	FROM orders
	where user_id = $1;

-- name: GetOrAddOrder :one
WITH inserted AS (
	INSERT INTO public.orders(order_num, status, user_id)
		VALUES ($1, $2, $3)
	ON CONFLICT (order_num) DO NOTHING
	RETURNING order_num, status, user_id, created_at, updated_at
)
SELECT order_num, status, user_id, created_at, updated_at, true as is_new
FROM inserted
UNION ALL
SELECT order_num, status, user_id, created_at, updated_at, false as is_new
FROM orders
WHERE order_num = $1
  AND NOT EXISTS (SELECT 1 FROM inserted);

-- name: UpdateOrder :one
UPDATE public.orders
	SET status = $2, accrual = $3, updated_at = now()
	WHERE order_num = $1
	RETURNING order_num, status, accrual, user_id;