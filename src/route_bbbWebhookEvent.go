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
	"BbbStatus/internal/BBBEvents"
	"encoding/json"
	"fmt"
	"github.com/jackc/pgx/v5"
	"github.com/labstack/echo/v4"
	"net"
	"net/http"
	"strings"
)

/*
bbbWebHookEvent is an HTTP handler function that processes incoming webhook events from a Big Blue Button (BBB) server.

The function performs the following steps:

1. Retrieves the "event" parameter from the incoming HTTP request form data.
2. Trims the leading and trailing square brackets from the event string.
3. Unmarshals the event string into a `BBBEvents.BaseEvent` struct.
4. Connects to the PostgreSQL database using the connection string specified in the `DB_CONNECTION_STRING` environment variable.
5. Calls the `Save` method on the `BBBEvents.BaseEvent` struct to persist the event data to the database.
6. Returns a 200 OK HTTP response.

Parameters:
- c (echo.Context): The Echo framework context, which provides access to the incoming HTTP request and response.

Returns:
- error: Any error that occurred during the processing of the webhook event.

Note:
- The function assumes that the `BBBEvents.BaseEvent` struct has a `Save` method that handles the persistence of the event data to the database.
- The function does not handle any errors that may occur during the database connection or the `Save` operation. It simply logs the errors and returns them.
*/
func bbbWebHookEvent(c echo.Context) error {
	var event BBBEvents.BaseEvent
	postEvent := c.FormValue("event")
	postEvent = strings.TrimLeft(postEvent, "[")
	postEvent = strings.TrimRight(postEvent, "]")
	apiKey := c.Request().Header.Get("Authorization")
	apiKey = strings.TrimSpace(strings.Replace(apiKey, "Bearer ", "", -1))
	// fmt.Println("API key: '" + apiKey + "'")
	fmt.Println("DEBUG bbbWebHookEvent -> postEvent: ", postEvent)

	err = json.Unmarshal(
		[]byte(postEvent),
		&event,
	)
	if err != nil {
		fmt.Println("Error occurred during unmarshalling of the event: ", err)
		return c.String(http.StatusBadRequest, "Error occurred during unmarshalling of the event")
	}

	addr, err := net.LookupAddr(getIpFromContext(c).String())
	if err != nil || len(addr) == 0 {
		fmt.Println("Failed to obtain BBBServer hostname of host:", getIpFromContext(c))
		return c.String(http.StatusBadRequest, "Failed to obtain BBBServer hostname of host")
	}
	if len(addr) > 1 {
		fmt.Println("LookupAddr returned more then one addr:", addr)
	}
	if len(addr) != 0 {
		event.Data.Attributes.Meeting.BbbHostname = strings.TrimSpace(strings.TrimRight(addr[0], "."))
		//fmt.Println("DEBUG bbbWebHookEvent -> BBBServer: '" + event.Data.Attributes.Meeting.BbbHostname + "'")
	}
	conn, err := pgx.Connect(c.Request().Context(), confGet("DB_CONNECTION_STRING"))
	if err != nil {
		fmt.Println("error occurred during pgx connect (bbbWebHookEvent): ", err)
		return c.String(http.StatusInternalServerError, "Error occurred database connect (bbbWebHookEvent)")
	}
	//goland:noinspection ALL
	defer conn.Close(c.Request().Context())

	// Simple host based webhook safety, not really great but better than nothing at all.
	bbbServers := confGetServers(event.Data.Attributes.Meeting.BbbHostname)
	if len(bbbServers) < 1 {
		//fmt.Println("List of valid hosts: ", confGetServers(""))
		//fmt.Println("Failed to authenticate host: '" + event.Data.Attributes.Meeting.BbbHostname + "'")
		return c.String(http.StatusUnauthorized, "Unauthorized")
	} else {
		server := bbbServers[0]
		if server.SharedSecret != apiKey {
			//fmt.Println("DEBUG: updating bbb API key")
			//fmt.Println("DEBUG server.SharedSecret='" + server.SharedSecret + "' api_key='" + apiKey + "'")
			server.SharedSecret = apiKey
			err = confSetServer(c.Request().Context(), server)
			if err != nil {
				fmt.Println("error occurred during confSetServer: ", err)
			}
		}
	}

	err = event.Save(c.Request().Context(), conn)
	if err != nil {
		fmt.Println("error occurred during save event: ", err)
		return c.JSON(http.StatusInternalServerError, map[string]interface{}{"error": err.Error()})
	}

	return c.String(http.StatusOK, "")
}
