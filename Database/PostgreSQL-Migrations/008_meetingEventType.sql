-- +goose up
CREATE TYPE meeting_event_type AS ENUM (
    'meeting-created',
    'meeting-ended'
    );
-- +goose Down
DROP TYPE meeting_event_type;