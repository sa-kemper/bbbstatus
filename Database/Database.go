package Database

import "embed"

//go:embed PostgreSQL-Migrations/*.sql
var PostgresInitialUp embed.FS
