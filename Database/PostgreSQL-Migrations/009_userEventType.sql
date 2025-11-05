-- +goose Up
CREATE TYPE user_event_type AS ENUM (
    'user-joined',
    'user-left',
    'user-presenter-as-signed',
    'user-audio-voice-enabled',
    'user-audio-voice-disabled',
    'user-audio-muted',
    'user-audio-unmuted',
    'user-cam-broadcast-start',
    'user-cam-broadcast-end',
    'user-emoji-changed',
    'user-raise-hand-changed'
    );
-- +goose Down
DROP TYPE user_event_type;