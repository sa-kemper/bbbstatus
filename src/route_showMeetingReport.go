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
	"bbbstatus/internal/BBBEvents"
	db "bbbstatus/internal/database"
	"errors"
	"fmt"
	"github.com/jackc/pgx/v5"
	"github.com/labstack/echo/v4"
	"github.com/nicksnyder/go-i18n/v2/i18n"
	"golang.org/x/text/language"
	"net/http"
	"os"
)

func init() { // Add all messages that are related to this file into the localization bundle
	var msgs = []i18n.Message{
		{ID: "ErrorTitleApplicationTimeout", Other: "Timeout"},
		{ID: "ErrorParagraphApplicationTimeout", Other: "bbbstatus reached a timeout, please try again."},
	}
	FrontendTextMessages = append(FrontendTextMessages, msgs...)
	for _, m := range BBBEvents.UserEventTextRepresentation { // Add user events text representation to the language strings.
		FrontendTextMessages = append(FrontendTextMessages, m)
	}
}

func showMeetingReport(c echo.Context) error {
	var internalMeetingId = c.Param("id")
	var requestLanguage = c.Request().Header.Get("Accept-Language")
	var meetingExists bool
	var ctx = c.Request().Context()
	localizer = i18n.NewLocalizer(Bundle, requestLanguage, language.English.String())

	// Connect to the db using
	conn, err := pgx.Connect(ctx, confGet("DB_CONNECTION_STRING"))
	if err != nil {
		return fmt.Errorf("Unable to connect to database: %v\n", err)
	}

	dbQueries := db.New(conn)

	//goland:noinspection ALL
	defer conn.Close(ctx)

	// Query and parse meeting using the row.next and row.scan methode of pgx
	meetingExists, err = dbQueries.GetMeetingExistsByID(ctx, internalMeetingId)
	if !meetingExists || err != nil {
		fmt.Println(err)
		return c.Render(http.StatusNotFound, "notfound", nil)
	}

	report, err := GenerateWebReport(ctx, internalMeetingId)
	if err != nil {
		if errors.Is(err, os.ErrDeadlineExceeded) {
			return c.Render(http.StatusGatewayTimeout, "errorPage", frontendError{ErrorTitle: Translate("ErrorTitleApplicationTimeout"), ErrorParagraph: Translate("ErrorParagraphApplicationTimeout")})
		}
		fmt.Println(err)
		return err
	}

	return c.Render(http.StatusOK, "report", map[string]interface{}{"InternalMeetingID": internalMeetingId, "Report": report})
}
