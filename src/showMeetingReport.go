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
	"github.com/jackc/pgx/v5"
	"github.com/labstack/echo/v4"
	"github.com/nicksnyder/go-i18n/v2/i18n"
	"golang.org/x/text/language"
	"net/http"
)

func showMeetingReport(c echo.Context) error {
	var internalMeetingId = c.Param("id")
	var requestLanguage = c.Request().Header.Get("Accept-Language")
	var meetingExtists bool
	var ctx = context.Background()
	localizer = i18n.NewLocalizer(Bundle, requestLanguage, language.English.String())

	// Connect to the db using
	conn, err := pgx.Connect(ctx, confGet("DB_CONNECTION_STRING"))
	if err != nil {
		return fmt.Errorf("Unable to connect to database: %v\n", err)
	}
	//goland:noinspection ALL
	defer conn.Close(ctx)

	// Query and parse meeting using the row.next and row.scan methode of pgx
	err = conn.QueryRow(ctx, "SELECT TRUE FROM meetings WHERE internal_meeting_id = $1 LIMIT 1", internalMeetingId).Scan(&meetingExtists)
	if !meetingExtists || err != nil {
		fmt.Println(err)
		return c.Render(http.StatusNotFound, "notfound", nil)
	}

	report, err := GenerateWebReport(context.Background(), internalMeetingId)
	if err != nil {
		fmt.Println(err)
		return err
	}

	return c.Render(http.StatusOK, "report", map[string]interface{}{"InternalMeetingID": internalMeetingId, "Report": report})
}
