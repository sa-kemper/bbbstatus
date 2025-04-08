-- name: GetMeetingById :one
SELECT *
FROM meetings
WHERE internal_meeting_id = $1
LIMIT 1;