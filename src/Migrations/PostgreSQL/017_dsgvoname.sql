-- +goose Up
ALTER TABLE users
    ADD COLUMN dsgvo_name VARCHAR(255) NOT NULL default 'unnamed user';
-- +goose Down
ALTER TABLE users
    DROP COLUMN dsgvo_name;
