-- +goose Up
ALTER TABLE meetings
    ADD COLUMN participant_count INTEGER DEFAULT 0;
ALTER TABLE meetings
    ADD CONSTRAINT participant_count CHECK (meetings.participant_count >= 0);
-- +goose Down
ALTER TABLE meetings
    DROP COLUMN participant_count CASCADE;
