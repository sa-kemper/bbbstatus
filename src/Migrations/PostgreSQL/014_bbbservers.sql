-- +goose Up
CREATE TABLE bbb_servers
(
    hostname      TEXT PRIMARY KEY,
    api_key       TEXT NOT NULL DEFAULT 'nil',
    friendly_name TEXT NOT NULL,
    active        bool NOT NULL DEFAULT TRUE,
    api_port      TEXT NOT NULL DEFAULT 443
);
-- +goose Down
DROP TABLE bbb_servers;