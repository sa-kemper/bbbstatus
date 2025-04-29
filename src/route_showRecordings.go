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
	"bbbstatus/internal/BBBAPI"
	bbbstatus "bbbstatus/internal/database"
	"fmt"
	"github.com/labstack/echo/v4"
	"net/http"
	"slices"
	"time"
)

func showRecordings(context echo.Context) (err error) {
	var ctx = context.Request().Context()
	servers := confGetServers("")
	globalRecordings := make([]BBBAPI.Recording, 0)
	for _, server := range servers {
		timeout := time.Duration(server.APITimeout) * time.Second
		api := BBBAPI.API{
			Hostname:     server.Hostname,
			Port:         server.ApiPort,
			SharedSecret: server.SharedSecret,
			Timeout:      &timeout,
		}
		recordings, err := api.GetRecordings(ctx, BBBAPI.GetRecordingsParameters{}, bbbstatus.Meeting{})
		if err != nil {
			fmt.Println("error occurred while getting recordings: ", err)
			return err
		}
		globalRecordings = slices.Insert(globalRecordings, len(globalRecordings), recordings.Recording...)
	}
	return context.Render(http.StatusOK, "recordings", map[string]interface{}{"Recordings": struct{ Recording []BBBAPI.Recording }{Recording: globalRecordings}})
}
