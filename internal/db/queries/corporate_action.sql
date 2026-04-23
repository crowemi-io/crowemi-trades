-- name: UpsertCorporateAction :one
INSERT INTO app.corporate_action (
    account_id, corporate_actions_id, ca_type, ca_sub_type, initiating_symbol,
    initiating_original_cusip, target_symbol, target_original_cusip,
    declaration_date, expiration_date, record_date, payable_date, cash, old_rate,
    new_rate, last_synced_at, created_at, created_by, updated_at, updated_by
) VALUES (
    $1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13, $14, $15, $16,
    NOW(), 'system', NOW(), 'system'
)
ON CONFLICT (corporate_actions_id) DO UPDATE SET
    ca_type = EXCLUDED.ca_type,
    ca_sub_type = EXCLUDED.ca_sub_type,
    initiating_symbol = EXCLUDED.initiating_symbol,
    initiating_original_cusip = EXCLUDED.initiating_original_cusip,
    target_symbol = EXCLUDED.target_symbol,
    target_original_cusip = EXCLUDED.target_original_cusip,
    declaration_date = EXCLUDED.declaration_date,
    expiration_date = EXCLUDED.expiration_date,
    record_date = EXCLUDED.record_date,
    payable_date = EXCLUDED.payable_date,
    cash = EXCLUDED.cash,
    old_rate = EXCLUDED.old_rate,
    new_rate = EXCLUDED.new_rate,
    last_synced_at = EXCLUDED.last_synced_at,
    updated_at = NOW(),
    updated_by = 'system'
RETURNING *;

-- name: ListCorporateActionsByAccountID :many
SELECT * FROM app.corporate_action
WHERE account_id = $1
ORDER BY last_synced_at DESC
LIMIT $2 OFFSET $3;

-- name: GetLatestCorporateActionByAccountID :one
SELECT * FROM app.corporate_action
WHERE account_id = $1
ORDER BY last_synced_at DESC
LIMIT 1;

-- name: GetCorporateActionByID :one
SELECT * FROM app.corporate_action WHERE corporate_actions_id = $1;
