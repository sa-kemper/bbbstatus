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
	"errors"
	"fmt"
	"github.com/jackc/pgx/v5"
	"github.com/labstack/echo/v4"
	"net/http"
	"os"
	"strings"
	"time"
)

func downloadMeetingReport(c echo.Context) error {
	var internalMeetingId = c.Param("id")
	var meetingName string
	var meetingDate time.Time
	var report, err = GenerateCSVReport(c.Request().Context(), internalMeetingId)
	if err != nil {
		if errors.Is(err, os.ErrDeadlineExceeded) { // this function may execute for a considerable amount of time. It's not completely unreasonable to assume that it may exceed time limits.
			fmt.Println("Generate CSV Report Timed out.")
		}
		return c.Render(http.StatusInternalServerError, "error", frontendError{ErrorTitle: Translate("ErrorTitleApplicationTimeout"), ErrorParagraph: Translate("ErrorParagraphApplicationTimeout")})
	}

	conn, err := pgx.Connect(c.Request().Context(), confGet("DB_CONNECTION_STRING"))
	if err != nil {
		fmt.Println(err)
	}
	err = conn.QueryRow(c.Request().Context(), "SELECT create_time, name FROM meetings WHERE internal_meeting_id = $1", internalMeetingId).Scan(&meetingDate, &meetingName)
	if err != nil {
		fmt.Println(err)
	}
	meetingName = strings.ReplaceAll(strings.ReplaceAll(meetingName, " ", "-"), "'", "")
	c.Response().Header().Set("Content-Disposition", "attachment; filename="+fmt.Sprintf("bbbstatus-meeting-report-%s-%s.csv", meetingName, meetingDate.Format("2006-02-01")))
	return c.Blob(http.StatusOK, "text/csv", report)
}
