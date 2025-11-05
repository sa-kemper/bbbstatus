-- name: InsertBBBServer :exec
INSERT INTO bbb_servers (hostname, api_key, friendly_name, active, api_port)
VALUES ($1, $2, $3, $4, $5);

-- name: GetBBBServerByHostname :one
SELECT *
FROM bbb_servers
WHERE hostname = $1;

-- name: SetRecordingsCountForHostname :exec
UPDATE bbb_servers
SET recordings_count = $1
WHERE hostname = $2;

-- name: GetRecordingsCountForAllServer :one
SELECT SUM(bbb_servers.recordings_count) AS totalRecordings
FROM bbb_servers;

-- name: IncrementRecordingsCountForServer :exec
UPDATE bbb_servers
SET recordings_count = recordings_count + 1
WHERE hostname = $1;