-- +goose Up
ALTER TABLE users
    DROP CONSTRAINT users_role_check;
ALTER TABLE users
    ADD CONSTRAINT users_role_check CHECK (role IN ('MODERATOR', 'VIEWER', 'SYSTEM'));
INSERT INTO users (internal_user_id, external_user_id, name, role, is_guest, join_timestamp, leave_timestamp)
VALUES ('SYSTEM', 'SYSTEM', 'SYSTEM', 'SYSTEM', FALSE, now(), now());

-- +goose Down
ALTER TABLE users
    DROP CONSTRAINT users_role_check;
ALTER TABLE users
    ADD CONSTRAINT users_role_check CHECK ( role IN ('MODERATOR', 'VIEWER'));
DELETE
FROM users
WHERE internal_user_id = 'SYSTEM';
