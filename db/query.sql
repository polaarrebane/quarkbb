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

-- name: GetAllSessionsByUserId :many
SELECT * FROM sessions
WHERE user_id = $1;

-- name: CountClosedSessionByPublicID :one
SELECT count(*) FROM sessions
WHERE public_id = $1 AND status = 'closed';

-- name: CreateSession :one
INSERT INTO sessions (
    user_id, public_id, created_at, updated_at, status
) VALUES (
    $1, $2, $3, $4, 'active'
)
RETURNING id;

-- name: CloseSession :one
UPDATE sessions
SET
    status = 'closed',
    updated_at = now()
WHERE public_id = $1
RETURNING public_id;
