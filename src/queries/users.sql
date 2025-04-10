-- name: GetUserNameById :one
SELECT name
FROM users
WHERE internal_user_id = $1;

-- name: GetUserById :one
SELECT *
FROM users
WHERE internal_user_id = $1;

-- name: GetPresenterUserByMeetingID :one
SELECT *
FROM users
WHERE internal_user_id = (SELECT user_events.internal_user_id
                          FROM user_events
                          WHERE internal_meeting_id = $1
                          ORDER BY event_timestamp DESC
                          LIMIT 1);

-- name: LeaveUserByID :exec
UPDATE users
SET leave_timestamp = $1
WHERE internal_user_id = $2;

-- name: GetUserExistsByID :one
SELECT TRUE
FROM users
WHERE internal_user_id = $1;

-- name: InsertUser :exec
INSERT INTO users (internal_user_id, external_user_id, name, role, is_guest)
VALUES ($1, $2, $3, $4, $5);