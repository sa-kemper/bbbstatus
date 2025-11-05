-- +goose Up
CREATE TABLE meeting_events
(
    event_id            UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    internal_meeting_id VARCHAR(255) NOT NULL,
    event_type          VARCHAR(50)  NOT NULL,
    event_timestamp     TIMESTAMP    NOT NULL,
    FOREIGN KEY (internal_meeting_id) REFERENCES meetings (internal_meeting_id)
);
-- +goose Down
DROP TABLE meeting_events;