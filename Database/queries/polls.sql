-- name: GetPollResponsesByPollID :many
SELECT *
FROM poll_responses
WHERE poll_id = $1;

-- name: GetPollAnswersByPollID :many
SELECT answers
FROM polls
WHERE poll_id = $1;