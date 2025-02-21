-- +goose Up
ALTER TABLE users
    ADD COLUMN join_timestamp  TIMESTAMP NOT NULL DEFAULT now(),
    ADD COLUMN leave_timestamp TIMESTAMP DEFAULT NULL;
-- +goose Down
ALTER TABLE users
    DROP COLUMN join_timestamp,
    DROP COLUMN leave_timestamp;