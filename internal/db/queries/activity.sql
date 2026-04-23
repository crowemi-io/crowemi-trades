-- name: CreateActivity :one
INSERT INTO app.activity (
    account_id, activity_id, activity_type, transaction_time, type, price, qty,
    side, symbol, leaves_qty, cum_qty, date, net_amount, description,
    per_share_amount, order_id, order_status, status, is_option, options_income,
    symbol_derived, created_at, created_by, updated_at, updated_by
) VALUES (
    $1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13, $14, $15, $16, $17,
    $18, $19, $20, $21, NOW(), 'system', NOW(), 'system'
)
RETURNING *;

-- name: GetLatestActivityByAccountID :one
SELECT * FROM app.activity
WHERE account_id = $1
ORDER BY transaction_time DESC
LIMIT 1;

-- name: ListActivitiesByAccountID :many
SELECT * FROM app.activity
WHERE account_id = $1
ORDER BY transaction_time DESC
LIMIT $2 OFFSET $3;

-- name: GetActivityByAccountAndID :one
SELECT * FROM app.activity
WHERE account_id = $1 AND activity_id = $2;
