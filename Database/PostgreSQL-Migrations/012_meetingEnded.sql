-- +goose UP
ALTER TABLE meetings
    DROP COLUMN duration;
ALTER TABLE meetings
    DROP COLUMN active;
ALTER TABLE meetings
    ADD COLUMN meeting_ended TIMESTAMP DEFAULT now(),
    ALTER COLUMN meeting_ended SET DEFAULT NULL;
-- Set ended to now for existing values, and null for new values

-- +goose Down
ALTER TABLE meetings
    DROP COLUMN meeting_ended;
ALTER TABLE meetings
    ADD COLUMN duration INTEGER;
ALTER TABLE meetings
    ADD COLUMN active bool;
