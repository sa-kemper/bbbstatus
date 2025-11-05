-- +goose Up
CREATE TABLE meetings
(
    internal_meeting_id VARCHAR(255) PRIMARY KEY,
    external_meeting_id VARCHAR(255) NOT NULL,
    name                VARCHAR(255) NOT NULL,
    is_breakout         BOOLEAN DEFAULT FALSE,
    parent_id           VARCHAR(255),
    duration            INTEGER DEFAULT 0,
    create_time         TIMESTAMP    NOT NULL,
    moderator_pass      VARCHAR(50)  NOT NULL,
    viewer_pass         VARCHAR(50)  NOT NULL,
    record              BOOLEAN DEFAULT FALSE,
    voice_conf          VARCHAR(50),
    dial_number         VARCHAR(50),
    max_users           INTEGER DEFAULT 0,
    metadata            JSONB   DEFAULT '{}'::jsonb,
    bbbHostname         VARCHAR(255) NOT NULL,
    active              bool    DEFAULT TRUE
);

-- +goose Down
DROP TABLE meetings;