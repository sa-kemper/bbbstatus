-- name: GetUserEventsByMeetingID :many
SELECT *
FROM user_events
WHERE internal_meeting_id = $1;

-- name: GetUserIDsFromMeetingByMeetingID :many
SELECT DISTINCT internal_user_id
FROM user_events
WHERE internal_meeting_id = $1;