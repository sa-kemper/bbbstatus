-- name: GetMeetingById :one
SELECT *
FROM meetings
WHERE internal_meeting_id = $1
LIMIT 1;


-- name: GetMeetingsBetweenDates :many
SELECT *
FROM meetings
WHERE create_time BETWEEN $1 AND $2;

-- name: GetMeetings :many
SELECT *
FROM meetings;