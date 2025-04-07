-- name: GetUserNameById :one
SELECT name
FROM users
WHERE internal_user_id = $1;