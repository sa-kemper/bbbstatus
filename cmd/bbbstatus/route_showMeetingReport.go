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
	"bbbstatus/internal/config"
	"bbbstatus/locales"
	"bbbstatus/pkg/BBBEvents"
	"context"
	"errors"
	"fmt"
	"net/http"
	"os"

	database "bbbstatus/internal/database"

	"github.com/jackc/pgx/v5"
	"github.com/labstack/echo/v4"
	"github.com/nicksnyder/go-i18n/v2/i18n"
	"golang.org/x/text/language"
)

func init() { // Add all messages that are related to this file into the localization bundle
	var msgs = []i18n.Message{
		{ID: "ErrorTitleApplicationTimeout", Other: "timeout"},
		{ID: "ErrorParagraphApplicationTimeout", Other: "bbbstatus reached a timeout, please try again."},
		{ID: "MODERATOR", Other: "moderator"},
		{ID: "VIEWER", Other: "viewer"},
		{ID: "MeetingReportMeetingDetailsGDPRCheckbox", Other: "view report GDPR compliant"},
		{ID: "MeetingReportMeetingDetailsGDPRFormSubmit", Other: "change"},
		{ID: "ForceCloseMeetingButton", Other: "force close this meeting"},
		{ID: "ForceCloseModalTitle", Other: "Do you want to Force close the meeting: "},
		{ID: "ForceCloseModalDescription", Other: "force closing this meeting in for bbbstatus does result in some functionality breaking. please only use this feature if you actually know that the meeting has ended, and bbbstatus is reporting it as active. the closing of active meetings is not supported and not tested and it may have unexpected consequences as this is not part of bbbstatus."},
		{ID: "ForceCloseModalAlertTitle", Other: "This action is dangerous"},
		{ID: "ForceCloseModalAlertText", Other: "Undoing this action can only be done by a database administrator"},
		{ID: "Cancel", Other: "Cancel"},
		{ID: "ForceCloseModalContinue", Other: "Force close this meeting"},
		{ID: "ForceCloseModalWarningTitle", Other: "ATTENTION:"},
		{ID: "ForceCloseModalWarningText", Other: "This does not close the meeting in big blue button!"},
	}
	FrontendTextMessages = append(FrontendTextMessages, msgs...)
	for _, m := range BBBEvents.UserEventTextRepresentation { // Add user events text representation to the language strings.
		FrontendTextMessages = append(FrontendTextMessages, m)
	}
}

func showMeetingReport(c echo.Context) error {
	var cc, ok = c.(*config.CustomContext)
	if !ok {
		return errors.New("cannot cast config.CustomContext")
	}

	var internalMeetingId = c.Param("id")
	var requestLanguage = c.Request().Header.Get("Accept-Language")
	var meetingExists bool
	//locales.Localizer = i18n.NewLocalizer(locales.Bundle, requestLanguage, language.English.String())
	localizer := i18n.NewLocalizer(locales.Bundle, requestLanguage, language.English.String())
	var ctx = context.WithValue(c.Request().Context(), locales.Translator("Translator"), localizer)

	// Connect to the db using
	conn, err := pgx.Connect(ctx, cc.Config.DatabaseConfig.DatabaseConnectionString)
	if err != nil {
		return fmt.Errorf("unable to connect to database: %v", err)
	}
	//goland:noinspection ALL
	defer conn.Close(ctx)

	// Query and parse meeting using the row.next and row.scan methode of pgx
	err = conn.QueryRow(ctx, "SELECT TRUE FROM meetings WHERE internal_meeting_id = $1 LIMIT 1", internalMeetingId).Scan(&meetingExists)
	if !meetingExists || err != nil {
		if err != nil {
			fmt.Println("error occurred querying meetings from database:", err)
		}
		return c.Render(http.StatusNotFound, "notfound", nil)
	}
	db := database.New(conn)
	var isMeetingActive bool
	meeting, queryErr := db.GetMeetingById(ctx, internalMeetingId)
	if queryErr != nil {
		fmt.Println("error occurred querying meetings from database:", err)
		isMeetingActive = false
	}
	isMeetingActive = meeting.MeetingEnded.Time.IsZero()

	report, err := GenerateWebReport(context.WithValue(ctx, Gdpr("gdpr"), c.QueryParam("gdpr") == "on"), cc.Config, internalMeetingId, nil)
	if err != nil {
		if errors.Is(err, os.ErrDeadlineExceeded) {
			fmt.Println("FATAL: database timeout!")
			return c.Render(http.StatusGatewayTimeout, "errorPage", frontendError{ErrorTitle: locales.TranslateFromEchoContext(c, "ErrorTitleApplicationTimeout"), ErrorParagraph: locales.TranslateFromEchoContext(c, "ErrorParagraphApplicationTimeout")})
		}
		fmt.Println(err)
		return err
	}

	return c.Render(http.StatusOK, "report", map[string]interface{}{"InternalMeetingID": internalMeetingId, "Report": report, "Gdpr": c.QueryParam("gdpr") == "on", "MeetingActive": isMeetingActive, "Meeting": meeting})
}
