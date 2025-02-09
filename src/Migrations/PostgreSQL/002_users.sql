-- +goose Up
CREATE TABLE users
(
    internal_user_id VARCHAR(255) PRIMARY KEY,
    external_user_id VARCHAR(255) NOT NULL,
    name             VARCHAR(255) NOT NULL,
    role             VARCHAR(50)  NOT NULL CHECK (role IN ('MODERATOR', 'VIEWER')),
    is_guest         BOOLEAN DEFAULT FALSE
);
-- +goose Down
DROP TABLE users;
