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
	db "bbbstatus/internal/database"
	"bbbstatus/locales"
	"errors"
	"fmt"
	"net/http"
	"os"
	"strings"

	"github.com/jackc/pgx/v5"
	"github.com/labstack/echo/v4"
)

func downloadMeetingReport(c echo.Context) error {
	var internalMeetingId = c.Param("id")
	var ctx = c.Request().Context()
	var report, err = GenerateCSVReport(ctx, internalMeetingId, nil)
	if err != nil {
		if errors.Is(err, os.ErrDeadlineExceeded) { // this function may execute for a considerable amount of time. It's not completely unreasonable to assume that it may exceed time limits.
			fmt.Println("Generate CSV Report Timed out.")
		}
		return c.Render(http.StatusInternalServerError, "error", frontendError{ErrorTitle: locales.TranslateFromEchoContext(c, "ErrorTitleApplicationTimeout"), ErrorParagraph: locales.TranslateFromEchoContext(c, "ErrorParagraphApplicationTimeout")})
	}

	conn, err := pgx.Connect(ctx, confGet("DB_CONNECTION_STRING"))
	if err != nil {
		fmt.Println(err)
	}
	defer conn.Close(ctx)
	dbQueries := db.New(conn)

	meeting, err := dbQueries.GetMeetingById(ctx, internalMeetingId)
	if err != nil {
		fmt.Println(err)
	}

	if !meeting.CreateTime.Valid {
		fmt.Println("WARNING: meeting create time is invalid, looks like a data inconsistency.")
	}

	meeting.Name = strings.ReplaceAll(strings.ReplaceAll(meeting.Name, " ", "-"), "'", "")
	c.Response().Header().Set("Content-Disposition", "attachment; filename="+fmt.Sprintf("bbbstatus-meeting-report-%s-%s.csv", meeting.Name, meeting.CreateTime.Time.Format("2006-02-01")))
	return c.Blob(http.StatusOK, "text/csv", report)
}
