-- +goose Up
ALTER TABLE users
    ADD COLUMN gdpr_name VARCHAR(255) NOT NULL default 'unnamed user';
-- +goose Down
ALTER TABLE users
    DROP COLUMN gdpr_name;
