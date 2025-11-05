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
	"fmt"
	"strings"

	"github.com/jackc/pgx/v5"
)

// confGetServers is a helper function to query the currently configured bbbServer's
func confGetServers(query string) (matches []bbbServer) {
	query = strings.TrimSpace(query)
	for _, server := range runtimeBbbServers {
		if strings.Contains(server.Hostname, query) {
			matches = append(matches, server)
		}
	}
	//fmt.Println("DEBUG confGetServers -> len=", len(matches), "matches=", matches)
	return
}

func confSetServer(ctx context.Context, updatedServer bbbServer) error {
	//fmt.Println("DEBUG, runtimeBbbServers:", runtimeBbbServers)
	//fmt.Println("DEBUG, configured bbb servers:", os.Getenv("BBB_SERVERS"))
	if len(updatedServer.Hostname) < 3 {
		panic("INVALID BBB_SERVERS on confSetServer call")
	}

	for index, runtimeServer := range runtimeBbbServers {
		if runtimeServer.Hostname == updatedServer.Hostname {
			var dbServer bbbServer
			if runtimeServer == updatedServer {
				fmt.Println("[WARNING] confSetServers called without a mutated bbbServer object.")
			}
			runtimeBbbServers[index] = updatedServer
			conn, err := pgx.Connect(ctx, confGet("DB_CONNECTION_STRING"))
			if err != nil {
				return err
			}

			err = conn.QueryRow(ctx, "SELECT api_port, api_key, friendly_name FROM bbb_servers WHERE hostname = $1", updatedServer.Hostname).Scan(&dbServer.ApiPort, &dbServer.SharedSecret, &dbServer.FriendlyName)
			if err != nil {
				//fmt.Println("DEBUG updatedServer ("+updatedServer.Hostname+")", updatedServer)

				_ = conn.Close(ctx)
				return err
			}
			//fmt.Println("DEBUG: dbServer:", dbServer)
			//fmt.Println("DEBUG: updatedServer:", updatedServer)
			if !updatedServer.isEqual(dbServer) { // is the dbServer outdated?
				_, err = conn.Exec(ctx, "UPDATE bbb_servers SET api_port = $1, api_key = $2,  friendly_name = $3 WHERE hostname = $4", updatedServer.ApiPort, updatedServer.SharedSecret, updatedServer.FriendlyName, updatedServer.Hostname)
				if err != nil {
					fmt.Println("Error occurred updating bbb_server ("+updatedServer.Hostname+"):", err)
					_ = conn.Close(ctx)
					return err
				}
			}
			_ = conn.Close(ctx)
			return nil
		}
	}
	runtimeBbbServers = append(runtimeBbbServers, updatedServer)
	return nil
}
