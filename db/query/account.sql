-- name: CreateAccount :one
INSERT INTO account (
  owner, balance, currency
) VALUES (
  $1, $2, $3
)
RETURNING *;

-- name: GetAccount :one
SELECT * FROM account
WHERE id = $1
LIMIT 1;

-- name: GetAccountForUpdate :one
SELECT * FROM account
WHERE id = $1
LIMIT 1
FOR NO KEY UPDATE;

-- name: ListAccounts :many
SELECT * FROM account
LIMIT $1 OFFSET $2;

-- name: AddAccountBalance :one
UPDATE account
SET balance = balance + sqlc.arg(amount)
WHERE id = sqlc.arg(id)
RETURNING id;

-- name: UpdateBalance :one
UPDATE account
SET balance = $2
WHERE id = $1
RETURNING id;

-- name: DeleteAccount :exec
DELETE FROM account
WHERE id = $1;