-- name: GetUserEventsByMeetingID :many
SELECT *
FROM user_events
WHERE internal_meeting_id = $1;

-- name: GetUserIDsFromMeetingByMeetingID :many
SELECT DISTINCT internal_user_id
FROM user_events
WHERE internal_meeting_id = $1;

-- name: InsertUserEvent :exec
INSERT INTO user_events (internal_meeting_id, internal_user_id, event_type, event_timestamp)
VALUES ($1, $2, $3, $4);