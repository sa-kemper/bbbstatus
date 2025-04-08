-- name: GetUserEventsByMeetingID :many
SELECT event_timestamp, internal_user_id, event_type
FROM user_events
WHERE internal_meeting_id = $1;

-- name: GetUserIDsFromMeetingByMeetingID :many
SELECT DISTINCT internal_user_id
FROM user_events
WHERE internal_meeting_id = $1;