-- name: GetMeetingMessagesByID :many
SELECT internal_user_id, message_content, send_time
FROM chat_messages
WHERE internal_meeting_id = $1;