-- name: GetMeetingEventsByInternalMeetingID :many
SELECT event_type, event_timestamp
FROM meeting_events
WHERE internal_meeting_id = $1;