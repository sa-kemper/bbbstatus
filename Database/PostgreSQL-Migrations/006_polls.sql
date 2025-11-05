-- +goose Up
CREATE TABLE polls
(
    poll_id             VARCHAR(255) PRIMARY KEY,
    internal_meeting_id VARCHAR(255) NOT NULL,
    internal_user_id    VARCHAR(255) NOT NULL,
    question            TEXT         NOT NULL,
    answers             JSONB        NOT NULL,
    created_at          TIMESTAMP    NOT NULL DEFAULT CURRENT_TIMESTAMP UNIQUE,
    FOREIGN KEY (internal_meeting_id) REFERENCES meetings (internal_meeting_id),
    FOREIGN KEY (internal_user_id) REFERENCES users (internal_user_id)
);
-- +goose Down
DROP TABLE polls;