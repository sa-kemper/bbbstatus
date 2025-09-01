-- name: GetFirstMeetingDate :one
SELECT min(meetings.create_time)
from meetings;
-- name: GetLastMeetingDate :one
SELECT max(meetings.meeting_ended)
from meetings;

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

-- name: GetMeetingExistsByID :one
SELECT TRUE
FROM meetings
WHERE internal_meeting_id = $1
LIMIT 1;

-- name: EndMeetingAtTimestampByID :exec
UPDATE meetings
SET meeting_ended = $1
WHERE internal_meeting_id = $2;

-- name: InsertMeeting :exec
INSERT INTO meetings (internal_meeting_id, external_meeting_id, name, is_breakout, parent_id, create_time,
                      moderator_pass, viewer_pass, record, voice_conf, dial_number, max_users, metadata, bbbhostname)
VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13, $14);

-- name: GetMeetingCountBetweenDates :one
SELECT COUNT(*)
FROM meetings
WHERE create_time BETWEEN $1 AND $2;

-- name: GetMeetingActiveByID :one
SELECT COALESCE(
               EXISTS (SELECT 1
                       FROM meetings
                       WHERE meeting_ended IS NULL
                         AND internal_meeting_id = $1),
               FALSE
       );