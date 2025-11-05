-- +goose Up
CREATE TABLE user_events
(
    event_id            UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    internal_meeting_id VARCHAR(255) NOT NULL,
    internal_user_id    VARCHAR(255) NOT NULL,
    event_type          VARCHAR(50)  NOT NULL,
    event_timestamp     TIMESTAMP    NOT NULL,
    FOREIGN KEY (internal_meeting_id) REFERENCES meetings (internal_meeting_id),
    FOREIGN KEY (internal_user_id) REFERENCES users (internal_user_id)
);
-- +goose Down
DROP TABLE user_events;
