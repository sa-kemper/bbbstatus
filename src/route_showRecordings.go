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
	"github.com/jackc/pgx/v5"
	"github.com/labstack/echo/v4"
	"github.com/lithammer/fuzzysearch/fuzzy"
	"github.com/nicksnyder/go-i18n/v2/i18n"
	"net/http"
	"slices"
	"sort"
	"strings"
	"time"
)

func init() {
	var msgs = []i18n.Message{
		{ID: "RecordingsPageTitle", Other: "Meeting Recordings"},
		{ID: "RecordingsPageSearchBar", Other: "Search for recording"},
		{ID: "RecordingPageRecordingState-published", Other: "Published"},
		{ID: "RecordingPageRecordingState-unpublished", Other: "Unpublished"},
		{ID: "RecordingPageRecordingMeetingID", Other: "Meeting ID"},
		{ID: "RecordingPageRecordingDuration", Other: "Duration"},
		{ID: "RecordingPageRecordingType", Other: "Type"},
		{ID: "RecordingPageRecordingBreakout", Other: "Breakout"},
		{ID: "RecordingPageNoRecordingsFound", Other: "No recordings found"},
		{ID: "RecordingPageNoRecordingsFoundHint", Other: "Try adjusting your search or filters to find recordings"},
		{ID: "RecordingPageRecordingPlay", Other: "Play"},
	}
	FrontendTextMessages = append(FrontendTextMessages, msgs...)
}

func showRecordings(context echo.Context) (err error) {
	var ctx = context.Request().Context()
	servers := confGetServers("")
	userQuery := context.QueryParam("query")
	userStartDate := context.FormValue("start-date")
	userEndDate := context.FormValue("end-date")

	// Get ALL Recordings
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
	//convert all meetings into a custom format that allows us to render time in the template.
	for iterator, recording := range globalRecordings {
		if recording.StartDate.IsZero() {
			globalRecordings[iterator].StartDate = time.UnixMilli(recording.StartTime)
		}
		if recording.EndDate.IsZero() {
			globalRecordings[iterator].EndDate = time.UnixMilli(recording.EndTime)
		}
	}

	// filter recordings based on date (start date specified by the user)
	if userStartDate != "" {
		startDate, err := time.Parse("2006-01-02", userStartDate)
		if err == nil {
			filteredRecordings := make([]BBBAPI.Recording, 0)
			for _, recording := range globalRecordings {
				if recording.StartDate.Before(startDate) {
					continue
				}
				filteredRecordings = append(filteredRecordings, recording)
			}
			globalRecordings = filteredRecordings
		}
		fmt.Println("error occurred while parsing start date: ", err)
	}

	// filter recordings based on date (end date specified by the user)
	if userEndDate != "" {
		endDate, err := time.Parse("2006-01-02", userEndDate)
		if err == nil {
			filteredRecordings := make([]BBBAPI.Recording, 0)
			for _, recording := range globalRecordings {
				if recording.StartDate.After(endDate) {
					continue
				}
				filteredRecordings = append(filteredRecordings, recording)
			}
			globalRecordings = filteredRecordings
		} else {
			fmt.Println("error occurred while parsing end date: ", err)
		}
	}

	// find other information that the user is looking for
	if userQuery != "" {
		conn, err := pgx.Connect(ctx, confGet("DB_CONNECTION_STRING"))
		if err != nil {
			_ = context.Render(http.StatusInternalServerError, "errorPage", map[string]interface{}{"ErrorTitle": "Internal Error", "ErrorParagraph": err.Error()})
			return err
		}
		defer conn.Close(ctx)

		dbQueries := bbbstatus.New(conn)

		var matchedRecordings []BBBAPI.Recording
		for _, recording := range globalRecordings {
			pariticipants, err := dbQueries.GetUsersInMeetingByInternalID(ctx, recording.MeetingID)
			if err != nil {
				_ = context.Render(http.StatusInternalServerError, "errorPage", map[string]interface{}{"ErrorTitle": err.Error()})
			}
			var participantNames = make([]string, len(pariticipants))
			for iterator, participant := range pariticipants {
				participantNames[iterator] = participant.Name
			}

			var fuzzyMatches = []bool{
				fuzzy.Match(userQuery, recording.RecordID),
				fuzzy.Match(userQuery, recording.MeetingID),
				fuzzy.Match(userQuery, recording.Name),
				fuzzy.Match(userQuery, strings.Join(participantNames, " ")),
			}
			if slices.Contains(fuzzyMatches, true) {
				matchedRecordings = append(matchedRecordings, recording)
			}
		}
		globalRecordings = matchedRecordings

	}

	sort.Slice(globalRecordings, func(i, j int) bool {
		return globalRecordings[i].StartDate.After(globalRecordings[j].StartDate)
	})

	err = context.Render(
		http.StatusOK,
		"recordings",
		map[string]interface{}{
			"Recordings": struct{ Recording []BBBAPI.Recording }{Recording: globalRecordings},
			"Request": struct {
				StartDate string
				EndDate   string
				Query     string
			}{
				StartDate: userStartDate,
				EndDate:   userEndDate,
				Query:     userQuery,
			},
		},
	)
	return err
}
