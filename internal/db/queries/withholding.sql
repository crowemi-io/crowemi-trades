-- name: CreateWithholding :one
INSERT INTO app.withholding (
    account_id, symbol, tax_type, amount, currency, occurred_at, description,
    created_at, created_by, updated_at, updated_by
) VALUES (
    $1, $2, $3, $4, $5, $6, $7, NOW(), 'system', NOW(), 'system'
)
RETURNING *;

-- name: ListWithholdingsByAccountID :many
SELECT * FROM app.withholding
WHERE account_id = $1
ORDER BY occurred_at DESC
LIMIT $2 OFFSET $3;

-- name: GetWithholdingByID :one
SELECT * FROM app.withholding WHERE id = $1;
