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
	"errors"
	"fmt"

	"github.com/jackc/pgx/v5"
)

func addBBBServerToDb(conn *pgx.Conn, server bbbServer) (err error) {
	var dbServer bbbServer
	var dbInitialized bool
	err = conn.QueryRow(context.TODO(), "SELECT EXISTS (SELECT 1 FROM information_schema.tables WHERE table_name = 'bbb_servers')").Scan(&dbInitialized)
	if err != nil {
		fmt.Println("FATAL ERROR occurred whilst checking if the database is initialized:", err)
		panic(err)
	}
	if !dbInitialized {
		// use a fallback if the db is not ready.
		if len(server.Hostname) < 3 {
			fmt.Println("addBBBServerToDb->server.Hostname='" + server.Hostname + "'")
			return errors.New("BBB_SERVERS hostname is too short")
		}
		runtimeBbbServers = append(runtimeBbbServers, bbbServer{Hostname: server.Hostname, ApiPort: server.ApiPort, SharedSecret: server.SharedSecret})
		return
	}

	// query the db for more information about the server.
	err = conn.QueryRow(context.TODO(), "SELECT api_port, api_key, friendly_name FROM bbb_servers WHERE hostname = $1 AND active = TRUE", server.Hostname).Scan(&dbServer.ApiPort, &dbServer.SharedSecret, &dbServer.FriendlyName)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			// Server was never used before
			//fmt.Println("DEBUG: Inserting " + server.Hostname + " to the database")
			_, err = conn.Exec(context.TODO(), "INSERT INTO bbb_servers (hostname, api_port, friendly_name) VALUES ($1, $2, $3)", server.Hostname, server.ApiPort, server.Hostname)
			if err != nil {
				fmt.Println("Error inserting bbb_server ("+server.Hostname+"):", err)
				panic(err)
			}
		} else {
			fmt.Println("Error occurred querying for the BBB server:", err)
		}
	}
	if len(dbServer.Hostname) < 3 {
		//fmt.Println("DEBUG: addBBBServerToDb->dbServer dbServer.Hostname='" + server.Hostname + "' dbServer.SharedSecret='" + server.SharedSecret + "'")
		runtimeBbbServers = append(runtimeBbbServers, server)
		//return errors.New("BBB_SERVERS hostname is too short")
	}
	runtimeBbbServers = append(runtimeBbbServers, dbServer)
	return
}
