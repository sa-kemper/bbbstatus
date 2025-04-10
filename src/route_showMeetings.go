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
	"github.com/jackc/pgx/v5/pgtype"
	"github.com/labstack/echo/v4"
	"github.com/nicksnyder/go-i18n/v2/i18n"
	"golang.org/x/text/language"
	"net/http"
	"os"
	"sort"
	"time"
)

func init() { // Add all messages that are related to this file into the localization bundle
	var msgs = []i18n.Message{
		{ID: "meetingListHeader", Other: "Meeting List"},
		{ID: "MeetingActive", Other: "Active"},
		{ID: "MeetingEnded", Other: "Ended"},
		{ID: "MeetingCreated", Other: "Created"},
		{ID: "MeetingToReport", Other: "Go to report"},
		{ID: "MeetingOpenDetails", Other: "Open details"},
		{ID: "MeetingExternalID", Other: "Internal ID"},
		{ID: "MeetingMaxUsers", Other: "Max Users"},
		{ID: "MeetingCreateTime", Other: "Create Time"},
		{ID: "MeetingDuration", Other: "Duration"},
		{ID: "MeetingRecordingStatusHeader", Other: "Recording"},
		{ID: "MeetingRecordingEnabled", Other: "Enabled"},
		{ID: "MeetingRecordingDisabled", Other: "Disabled"},
		{ID: "MeetingNoMeetingsExist", Other: "No Meetings Exist"},
		{ID: "meetingListHeader", Other: "Meeting List"},
		{ID: "meetingListFilterStartDate", Other: "Start Date"},
		{ID: "meetingListFilterEndDate", Other: "End Date"},
		{ID: "meetingListFilterFilterButton", Other: "Filter"},
		{ID: "meetingListHeaderSeeStatisticsButton", Other: "See statistics"},
	}
	FrontendTextMessages = append(FrontendTextMessages, msgs...)
}

// MeetingListMeetingWrapper is a container for BBBEvents.MeetingListMeetingWrapper.
// I need this so, I can add a meeting URL parameter because I can't get the reverse function inside the template to work.
type MeetingListMeetingWrapper struct {
	BBBEventsMeeting BBBEvents.Meeting
	BbbHostname      string
	Active           bool
}

// e.Get("/meetings/", showMeetings)
func showMeetings(c echo.Context) error {
	var isFilteredRequest bool
	var startDate, endDate time.Time
	var requestLanguage = c.Request().Header.Get("Accept-Language")
	var ctx = c.Request().Context()
	localizer = i18n.NewLocalizer(Bundle, requestLanguage, language.English.String())

	// handle filtered request
	startDateParam := c.FormValue("start-date")
	endDateParam := c.FormValue("end-date")
	if startDateParam != "" || endDateParam != "" {
		isFilteredRequest = true
	}

	startDate, err = time.Parse("2006-01-02", startDateParam)
	endDate, err = time.Parse("2006-01-02", endDateParam)

	conn, err := pgx.Connect(ctx, confGet("DB_CONNECTION_STRING"))
	if err != nil {
		return err
	}
	dbQueries := db.New(conn)

	//goland:noinspection ALL
	defer conn.Close(ctx)
	var meetings []MeetingListMeetingWrapper
	if isFilteredRequest {
		if endDate.IsZero() {
			endDate = time.Now()
		}
		endDate = endDate.Add(time.Hour * 25) // make the end date inclusive.
		dbMeetings, err := dbQueries.GetMeetingsBetweenDates(ctx, db.GetMeetingsBetweenDatesParams{
			CreateTime: pgtype.Timestamp{
				Valid: true,
				Time:  startDate,
			},
			CreateTime_2: pgtype.Timestamp{
				Valid: true,
				Time:  endDate,
			},
		})
		if err != nil {
			if errors.Is(err, os.ErrDeadlineExceeded) {
				fmt.Println("FATAL: Database timed out")
			}
			fmt.Println("Error occurred in GetMeetingsBetweenDates: ", err)
			return err
		}

		for _, m := range dbMeetings {
			rm := BBBEvents.ConvertDBToBBBMeeting(m)
			meetings = append(meetings, MeetingListMeetingWrapper{BBBEventsMeeting: rm, BbbHostname: rm.BbbHostname, Active: m.MeetingEnded.Time.Before(time.Now())})
		}
	} else {
		dbMeetings, err := dbQueries.GetMeetings(ctx)
		if err != nil {
			if errors.Is(err, os.ErrDeadlineExceeded) {
				fmt.Println("FATAL: Database timed out")
			}
			fmt.Println("Error getting meetings", err.Error())
			return err
		}
		for _, m := range dbMeetings {
			rm := BBBEvents.ConvertDBToBBBMeeting(m)
			meetings = append(meetings, MeetingListMeetingWrapper{BBBEventsMeeting: rm, BbbHostname: rm.BbbHostname, Active: m.MeetingEnded.Time.Before(time.Now())})
		}
	}

	sort.Slice(meetings, func(i, j int) bool {
		return meetings[i].BBBEventsMeeting.CreateTime.After(meetings[j].BBBEventsMeeting.CreateTime)
	})

	err = c.Render(http.StatusOK, "meetings", map[string]interface{}{"Request": struct {
		StartDate string
		EndDate   string
	}{StartDate: startDateParam, EndDate: endDateParam}, "Meetings": meetings})
	if err != nil {
		fmt.Println(err)
	}
	return nil
}
