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
	"bbbstatus/internal/apiCredentialHelper"
	"bbbstatus/internal/config"
	db "bbbstatus/internal/database"
	"bbbstatus/locales"
	"bbbstatus/pkg/BBBEvents"
	"context"
	"encoding/json"
	"fmt"
	"net"
	"net/http"
	"strings"

	"github.com/jackc/pgx/v5"
	"github.com/labstack/echo/v4"
)

/*
bbbWebHookEvent is an HTTP handler function that processes incoming webhook events from a Big Blue Button (BBB) server.

The function performs the following steps:

1. Retrieves the "event" parameter from the incoming HTTP request form data.
2. Trims the leading and trailing square brackets from the event string.
3. Unmarshals the event string into a `BBBEvents.BaseEvent` struct.
4. Connects to the PostgreSQL-Migrations database using the connection string specified in the `DB_CONNECTION_STRING` environment variable.
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
func bbbWebHookEvent(c echo.Context) (err error) {
	cc := c.(*config.CustomContext)
	// Debug deployments
	if cc.Config.BaseConfig.ClearQueue {
		return c.String(http.StatusOK, "Event was dropped.")
	}
	var ctx = context.WithValue(c.Request().Context(), locales.ServerLanguage("SERVER_LANG"), cc.Config.BaseConfig.ServerLang) // add a variable to the context in order to use it inside the BBBEvents package.
	var event BBBEvents.BaseEvent
	postEvent := c.FormValue("event")
	postEvent = strings.TrimLeft(postEvent, "[")
	postEvent = strings.TrimRight(postEvent, "]")
	apiKey := c.Request().Header.Get("Authorization")
	apiKey = strings.TrimSpace(strings.Replace(apiKey, "Bearer ", "", -1))
	// fmt.Println("API key: '" + apiKey + "'")
	// fmt.Println("DEBUG bbbWebHookEvent -> postEvent: ", postEvent)

	var requesterIpAddress string

	err = json.Unmarshal(
		[]byte(postEvent),
		&event,
	)
	if err != nil {
		fmt.Println("Error occurred during unmarshalling of the event: ", err)
		return c.String(http.StatusBadRequest, "Error occurred during unmarshalling of the event")
	}
	addr, err := authorizeBBBServer(c, requesterIpAddress)
	if err != nil {
		fmt.Println("Error occurred during authorizing of the bbbserver for event: with", postEvent, "error: ", err)
		return c.String(http.StatusBadRequest, "Error occurred during authorizing of the event")
	}

	if len(addr) > 1 {
		fmt.Println("LookupAddr returned more then one addr:", addr)
	}
	if len(addr) != 0 {
		event.Data.Attributes.Meeting.BbbHostname = strings.TrimSpace(strings.TrimRight(addr[0], "."))
		//fmt.Println("DEBUG bbbWebHookEvent -> BBBServer: '" + event.Data.Attributes.Meeting.BbbHostname + "'")
	}
	conn, err := pgx.Connect(ctx, cc.Config.DatabaseConfig.DatabaseConnectionString)
	if err != nil {
		fmt.Println("error occurred during pgx connect (bbbWebHookEvent): ", err)
		return c.String(http.StatusInternalServerError, "Error occurred database connect (bbbWebHookEvent)")
	}
	defer conn.Close(ctx)

	dbQueries := db.New(conn)

	// Simple host based webhook safety, not really great but better than nothing at all.
	if apiCredentialHelper.GetApiKey(event.Data.Attributes.Meeting.BbbHostname) != apiKey {
		err = apiCredentialHelper.UpdateApiKey(event.Data.Attributes.Meeting.BbbHostname, apiKey)
		if err != nil {
			fmt.Println("Error occurred during updating api key ( of server "+event.Data.Attributes.Meeting.BbbHostname+" ) err : ", err)
		}
	}

	err = event.Save(ctx, dbQueries, conn)
	if err != nil {
		fmt.Println("error occurred during save event: ", err)
		return c.Render(http.StatusInternalServerError, "BBBWebHookEvent", map[string]interface{}{})
	}

	return c.String(http.StatusOK, "")
}

func authorizeBBBServer(c echo.Context, requesterIpAddress string) ([]string, error) {
	frameworkExtractedIP := c.RealIP()
	customIpExtractor := getIpFromContext(c).String()

	// this is required due to differences with X86_64 and ARM implementations of the echo frameworks ip extractor.
	if ip := net.ParseIP(frameworkExtractedIP); ip != nil {
		requesterIpAddress = ip.String()
	}

	// this is required due to differences with X86_64 and ARM implementations of the echo frameworks ip extractor.
	if ip := net.ParseIP(customIpExtractor); ip != nil {
		requesterIpAddress = ip.String()
	}

	addr, err := net.LookupAddr(requesterIpAddress)
	if err != nil {
		fmt.Println("Failed to obtain BBBServer hostname of host:", requesterIpAddress)
		fmt.Println("Error:", err)
		return nil, c.String(http.StatusBadRequest, "Failed to obtain BBBServer hostname of host")

	}
	if len(addr) == 0 {
		fmt.Println("Failed to obtain BBBServer hostname of host (Zero addresses returned.):", requesterIpAddress)
		return nil, c.String(http.StatusBadRequest, "Failed to obtain BBBServer hostname of host")
	}
	return addr, err
}
