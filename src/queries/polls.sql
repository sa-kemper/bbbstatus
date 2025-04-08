-- name: GetPollResponsesByPollID :many
SELECT internal_user_id, answer_ids, response_time
FROM poll_responses
WHERE poll_id = $1;

-- name: GetPollAnswersByPollID :many
SELECT answers
FROM polls
WHERE poll_id = $1;