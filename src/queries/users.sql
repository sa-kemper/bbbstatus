-- name: GetUserNameById :one
SELECT name
FROM users
WHERE internal_user_id = $1;


-- name: GetUserById :one
SELECT *
FROM users
WHERE internal_user_id = $1;