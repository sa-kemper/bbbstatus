-- +goose Up
ALTER TABLE bbb_servers
    ADD COLUMN recordings_count int DEFAULT 0 NOT NULL;
-- +goose Down
ALTER TABLE bbb_servers
    DROP COLUMN recordings_count;
