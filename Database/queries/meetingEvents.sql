-- name: GetMeetingEventsByInternalMeetingID :many
SELECT *
FROM meeting_events
WHERE internal_meeting_id = $1;

-- name: InsertMeetingEventForID :exec
INSERT INTO meeting_events (internal_meeting_id, event_type, event_timestamp)
VALUES ($1, $2, $3);