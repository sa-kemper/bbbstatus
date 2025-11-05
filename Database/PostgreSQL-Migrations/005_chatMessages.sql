-- +goose Up
CREATE TABLE chat_messages
(
    message_id          UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    internal_meeting_id VARCHAR(255) NOT NULL,
    internal_user_id    VARCHAR(255) NOT NULL,
    chat_id             VARCHAR(50)  NOT NULL,
    message_content     TEXT         NOT NULL,
    send_time           TIMESTAMP    NOT NULL,
    FOREIGN KEY (internal_meeting_id) REFERENCES meetings (internal_meeting_id),
    FOREIGN KEY (internal_user_id) REFERENCES users (internal_user_id)
);
-- +goose Down
DROP TABLE chat_messages;