-- +goose Up
CREATE TABLE poll_responses
(
    response_id      UUID PRIMARY KEY      DEFAULT gen_random_uuid(),
    poll_id          VARCHAR(255) NOT NULL,
    internal_user_id VARCHAR(255) NOT NULL,
    answer_ids       JSONB        NOT NULL,
    response_time    TIMESTAMP    NOT NULL DEFAULT CURRENT_TIMESTAMP,
    FOREIGN KEY (poll_id) REFERENCES polls (poll_id),
    FOREIGN KEY (internal_user_id) REFERENCES users (internal_user_id)
);
-- +goose Down
DROP TABLE poll_responses;