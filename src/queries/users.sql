-- name: GetUserNameById :one
SELECT name
FROM users
WHERE internal_user_id = $1;


-- name: GetUserById :one
SELECT internal_user_id, external_user_id, name, role, is_guest
FROM users
WHERE internal_user_id = $1;