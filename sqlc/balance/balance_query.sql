-- name: GetWithdrawals :many
SELECT order_num, withdraw_amount, user_id, created_at
	FROM public.withdraw_history
	where user_id = $1;

-- name: GetBalance :one
SELECT amount, withdrawn, user_id
	FROM public.balance
	where user_id = $1;


-- name: CreateBalance :exec
INSERT INTO public.balance(user_id)
	VALUES ($1);

-- name: InsertWithdrawal :exec
INSERT INTO public.withdraw_history(order_num, withdraw_amount, user_id)
	VALUES ($1, $2, $3);