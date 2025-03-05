/*
 * Copyright 2025 Samuel Kemper
 *
 * Licensed under the Apache License, Version 2.0 (the "License");
 * you may not use this file except in compliance with the License.
 * You may obtain a copy of the License at
 *
 *     http://www.apache.org/licenses/LICENSE-2.0
 *
 * Unless required by applicable law or agreed to in writing, software
 * distributed under the License is distributed on an "AS IS" BASIS,
 * WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
 * See the License for the specific language governing permissions and
 * limitations under the License.
 */

package main

import (
	"database/sql"
	"embed"
	_ "embed"
	"fmt"
	_ "github.com/jackc/pgx/v5/stdlib"
	"github.com/pressly/goose/v3"
)

//go:embed Migrations/PostgreSQL/*.sql
var postgresInitialUp embed.FS

func initDatabase() error {
	db, err := sql.Open("pgx", confGet("DB_CONNECTION_STRING"))
	if err != nil {
		panic(fmt.Errorf("Unable to connect to database: %v\n", err))
	}
	defer db.Close()

	err = db.Ping()
	if err != nil {
		panic(fmt.Errorf("Unable to connect to database: %v\n", err))
	}

	err = goose.SetDialect("pgx")
	if err != nil {
		panic(fmt.Errorf("Unable to connect to database: %v\n", err))
	}
	goose.SetBaseFS(postgresInitialUp)

	if err := goose.Up(db, "Migrations/PostgreSQL"); err != nil { //
		panic(err)
	}
	if err := goose.Version(db, "Migrations/PostgreSQL"); err != nil {
		panic(err)
	}
	return nil
}
