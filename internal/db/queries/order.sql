-- name: UpsertOrder :one
INSERT INTO app.order (
    account_id, client_order_id, asset_id, symbol, side, type, order_class,
    time_in_force, status, quantity, notional, filled_qty, filled_avg_price,
    limit_price, stop_price, alpaca_created_at, alpaca_updated_at, submitted_at,
    filled_at, expired_at, canceled_at, created_at, created_by, updated_at, updated_by
) VALUES (
    $1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13, $14, $15, $16, $17,
    $18, $19, $20, $21, NOW(), 'system', NOW(), 'system'
)
ON CONFLICT (client_order_id) DO UPDATE SET
    asset_id = EXCLUDED.asset_id,
    symbol = EXCLUDED.symbol,
    side = EXCLUDED.side,
    type = EXCLUDED.type,
    order_class = EXCLUDED.order_class,
    time_in_force = EXCLUDED.time_in_force,
    status = EXCLUDED.status,
    quantity = EXCLUDED.quantity,
    notional = EXCLUDED.notional,
    filled_qty = EXCLUDED.filled_qty,
    filled_avg_price = EXCLUDED.filled_avg_price,
    limit_price = EXCLUDED.limit_price,
    stop_price = EXCLUDED.stop_price,
    alpaca_created_at = EXCLUDED.alpaca_created_at,
    alpaca_updated_at = EXCLUDED.alpaca_updated_at,
    submitted_at = EXCLUDED.submitted_at,
    filled_at = EXCLUDED.filled_at,
    expired_at = EXCLUDED.expired_at,
    canceled_at = EXCLUDED.canceled_at,
    updated_at = NOW(),
    updated_by = 'system'
RETURNING *;

-- name: GetLatestOrderByAccountID :one
SELECT * FROM app.order
WHERE account_id = $1
ORDER BY alpaca_created_at DESC
LIMIT 1;

-- name: ListOrdersByAccountID :many
SELECT * FROM app.order
WHERE account_id = $1
ORDER BY alpaca_created_at DESC
LIMIT $2 OFFSET $3;

-- name: GetOrderByClientOrderID :one
SELECT * FROM app.order WHERE client_order_id = $1;
