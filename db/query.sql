-- name: CountUserByUsername :one
SELECT count(*) FROM users
WHERE username = $1;

-- name: CreateUserAndReturnId :one
INSERT INTO users (
  username, email, password
) VALUES (
    $1, $2, $3
)
RETURNING id;

-- name: GetUserById :one
SELECT * FROM users
WHERE id = $1;

-- name: FindUserByUsername :one
SELECT * FROM users
WHERE username = $1;

-- name: CountUsedRefreshTokenByJTI :one
SELECT count(*) FROM used_refresh_tokens
WHERE jti = $1;

-- name: MarkTokenAsUsed :exec
INSERT INTO used_refresh_tokens (
    jti
) VALUES (
    $1
);
