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

-- name: GetUserCountBetweenDates :one
SELECT COUNT(*)
FROM users
WHERE join_timestamp BETWEEN $1 AND $2;

-- name: GetUsersWhoJoinedBetween :many
SELECT *
FROM users
WHERE join_timestamp BETWEEN $1 AND $2;

-- name: GetUserCountInMeetingByInternalID :one
SELECT count(internal_user_id) AS userCount
FROM users
WHERE internal_user_id IN (SELECT DISTINCT internal_user_id FROM user_events WHERE internal_meeting_id = $1);

-- name: GetUsersInMeetingByInternalID :many
SELECT internal_user_id AS users
FROM users
WHERE internal_user_id IN (SELECT DISTINCT internal_user_id FROM user_events WHERE internal_meeting_id = $1);