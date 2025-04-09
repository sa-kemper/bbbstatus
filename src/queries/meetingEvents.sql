-- name: GetMeetingEventsByInternalMeetingID :many
SELECT *
FROM meeting_events
WHERE internal_meeting_id = $1;