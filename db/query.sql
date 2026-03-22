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

-- name: FindUserByUsername :one
SELECT * FROM users
WHERE username = $1;
