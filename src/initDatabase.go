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
	"context"
	_ "embed"
	"errors"
	"fmt"
	"github.com/jackc/pgx/v5"
)

//go:embed Migrations/PostgreSQL/InitialUp.sql
var postgresInitialUp []byte

var RequiredTables = [7]string{"chat_messages", "meeting_events", "meetings", "polls", "poll_responses", "user_events", "users"}

func initDatabase() error {
	var ctx = context.Background()
	conn, err := pgx.Connect(ctx, confGet("DB_CONNECTION_STRING"))
	defer conn.Close(ctx)
	if err != nil {
		panic(fmt.Errorf("Unable to connect to database: %v\n", err))
	}
	err = conn.Ping(ctx)
	if err != nil {
		panic(fmt.Errorf("Unable to connect to database: %v\n", err))
	}

	for _, table := range RequiredTables {
		var tableExists bool
		err := conn.QueryRow(ctx, "SELECT EXISTS (SELECT 1 FROM information_schema.tables WHERE table_schema = 'public' AND table_name = $1)", table).Scan(&tableExists)
		if errors.Is(err, pgx.ErrNoRows) || !tableExists {
			fmt.Printf("Table %s does not exist\n", table)
			_, err = conn.Exec(ctx, string(postgresInitialUp))
			if err != nil {
				return err
			}
			return nil
		}
		if err != nil {
			return err
		}
	}
	return nil
}
