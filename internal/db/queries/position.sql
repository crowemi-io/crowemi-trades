-- name: UpsertPosition :one
INSERT INTO app.position (
    account_id, symbol, quantity, avg_entry_price, market_value, unrealized_pl,
    is_current, recorded_at, created_at, created_by, updated_at, updated_by
) VALUES (
    $1, $2, $3, $4, $5, $6, $7, $8, NOW(), 'system', NOW(), 'system'
)
ON CONFLICT (account_id, symbol) DO UPDATE SET
    quantity = EXCLUDED.quantity,
    avg_entry_price = EXCLUDED.avg_entry_price,
    market_value = EXCLUDED.market_value,
    unrealized_pl = EXCLUDED.unrealized_pl,
    is_current = EXCLUDED.is_current,
    recorded_at = EXCLUDED.recorded_at,
    updated_at = NOW(),
    updated_by = 'system'
RETURNING *;

-- name: ListCurrentPositionsByAccountID :many
SELECT * FROM app.position
WHERE account_id = $1 AND is_current = TRUE
ORDER BY symbol ASC;

-- name: ListPositionsByAccountID :many
SELECT * FROM app.position
WHERE account_id = $1
ORDER BY symbol ASC;

-- name: MarkPositionsStale :exec
UPDATE app.position
SET is_current = FALSE, updated_at = NOW(), updated_by = 'system'
WHERE account_id = $1 AND is_current = TRUE;

-- name: GetPositionByAccountAndSymbol :one
SELECT * FROM app.position WHERE account_id = $1 AND symbol = $2;
