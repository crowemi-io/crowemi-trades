-- name: CreatePortfolio :one
INSERT INTO app.portfolio (
    account_id, portfolio_key, name, weight, created_at, created_by, updated_at, updated_by
) VALUES (
    $1, $2, $3, $4, NOW(), 'system', NOW(), 'system'
)
RETURNING *;

-- name: GetPortfolioByID :one
SELECT * FROM app.portfolio WHERE id = $1;

-- name: GetPortfolioByAccountAndName :one
SELECT * FROM app.portfolio WHERE account_id = $1 AND name = $2;

-- name: GetPortfolioByAccount :many
SELECT * FROM app.portfolio WHERE account_id = $1;

-- name: ListPortfoliosByAccountID :many
SELECT * FROM app.portfolio WHERE account_id = $1 ORDER BY name ASC;

-- name: UpdatePortfolio :one
UPDATE app.portfolio
SET
    portfolio_key = $2,
    name = $3,
    weight = $4,
    updated_at = NOW(),
    updated_by = 'system'
WHERE id = $1
RETURNING *;

-- name: DeletePortfolio :exec
DELETE FROM app.portfolio WHERE id = $1;

-- name: CreatePortfolioSymbol :one
INSERT INTO app.portfolio_symbol (
    portfolio_id, symbol, weight, created_at, created_by, updated_at, updated_by
) VALUES (
    $1, $2, $3, NOW(), 'system', NOW(), 'system'
)
RETURNING *;

-- name: ListPortfolioSymbolsByAccountID :many
SELECT
    portfolio_symbol.symbol
FROM app.portfolio_symbol
    JOIN app.portfolio ON portfolio.id = portfolio_symbol.portfolio_id
WHERE portfolio.account_id = $1
ORDER BY symbol ASC;

-- name: DeletePortfolioSymbol :exec
DELETE FROM app.portfolio_symbol WHERE portfolio_id = $1 AND symbol = $2;
