-- name: GetMeetingMessagesByID :many
SELECT *
FROM chat_messages
WHERE internal_meeting_id = $1;