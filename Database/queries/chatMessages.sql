-- name: GetMeetingMessagesByID :many
SELECT *
FROM chat_messages
WHERE internal_meeting_id = $1;

-- name: InsertChatMessageToMeetingByID :exec
INSERT INTO chat_messages (internal_meeting_id, internal_user_id, chat_id, message_content, send_time)
VALUES ($1, $2, $3, $4, $5);