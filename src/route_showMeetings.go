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
	"bbbstatus/internal/BBBEvents"
	db "bbbstatus/internal/database"
	"context"
	"errors"
	"fmt"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgtype"
	"github.com/labstack/echo/v4"
	"github.com/nicksnyder/go-i18n/v2/i18n"
	"golang.org/x/text/language"
	"net/http"
	"os"
	"slices"
	"sort"
	"time"
)

func init() { // Add all messages that are related to this file into the localization bundle
	var msgs = []i18n.Message{
		{ID: "meetingListHeader", Other: "meeting list"},
		{ID: "MeetingActive", Other: "active"},
		{ID: "MeetingEnded", Other: "ended"},
		{ID: "MeetingCreated", Other: "created"},
		{ID: "MeetingToReport", Other: "go to report"},
		{ID: "MeetingOpenDetails", Other: "open details"},
		{ID: "MeetingExternalID", Other: "internal ID"},
		{ID: "MeetingUserCount", Other: "users"},
		{ID: "MeetingCreateTime", Other: "create Time"},
		{ID: "MeetingDuration", Other: "duration"},
		{ID: "MeetingRecordingStatusHeader", Other: "recording"},
		{ID: "MeetingRecordingEnabled", Other: "enabled"},
		{ID: "MeetingRecordingDisabled", Other: "disabled"},
		{ID: "MeetingNoMeetingsExist", Other: "no meetings exist"},
		{ID: "MeetingListHeader", Other: "meeting list"},
		{ID: "MeetingListFilterStartDate", Other: "start date"},
		{ID: "MeetingListFilterEndDate", Other: "end date"},
		{ID: "MeetingListFilterFilterButton", Other: "filter"},
		{ID: "MeetingListHeaderSeeStatisticsButton", Other: "see statistics"},
		{ID: "MeetingListUserCountLabel", One: "user", Other: "users"},
		{ID: "ServerStatsMeetings", One: "meeting", Other: "meetings"},
		{ID: "MeetingListHeaderSeeRecordingsButton", Other: "see recordings"},
		{ID: "MeetingListReloadButton", Other: "reload"},
	}
	FrontendTextMessages = append(FrontendTextMessages, msgs...)
}

// MeetingListMeetingWrapper is a container for BBBEvents.MeetingListMeetingWrapper.
// I need this so, I can add a meeting URL parameter because I can't get the reverse function inside the template to work.
type MeetingListMeetingWrapper struct {
	BBBEventsMeeting BBBEvents.Meeting
	BbbHostname      string
	UserCount        int
	Active           bool
}

// e.Get("/meetings/", showMeetings)
func showMeetings(c echo.Context) error {
	type ServerFilter struct {
		Hostname    string
		Users       int
		FilteredFor bool
	}
	var serverStats struct {
		TotalMeetings int
		ServerCounts  []struct {
			Percentage float32
			Hostname   string
			Meetings   int
		}
	}
	var isFilteredRequest bool
	var startDate, endDate time.Time
	var requestLanguage = c.Request().Header.Get("Accept-Language")
	var ctx = c.Request().Context()
	localizer = i18n.NewLocalizer(Bundle, requestLanguage, language.English.String())

	// handle filtered request
	startDateParam := c.FormValue("start-date")
	endDateParam := c.FormValue("end-date")

	var selectedServers []bbbServer
	var showMeetingsServerFiltered []ServerFilter

	for _, server := range confGetServers("") {
		apiTimeout := time.Duration(server.APITimeout) * time.Second
		BbbApi := BBBAPI.API{Hostname: server.Hostname, Port: server.ApiPort, SharedSecret: server.SharedSecret, Timeout: &apiTimeout}
		if valid, err := BbbApi.ValidateApiSettings(ctx); err != nil || !valid {
			if err != nil {
				fmt.Println("error occurred validating API settings for server '"+server.Hostname+"': ", err)
				continue
			}
			if !valid {
				fmt.Println(server.Hostname, "has no valid API settings")
				continue
			}
		}

		// fill server usage stats by meeting count
		serverMeetings, err := BbbApi.GetMeetings(ctx)
		if err != nil {
			fmt.Printf("Error getting meetings: %v\n", err)
			continue
		}
		serverStats.TotalMeetings += len(serverMeetings)
		serverStats.ServerCounts = append(serverStats.ServerCounts, struct {
			Percentage float32
			Hostname   string
			Meetings   int
		}{
			Percentage: 0.0,
			Hostname:   server.Hostname,
			Meetings:   len(serverMeetings),
		})

		userCount, err := BbbApi.GetServerUserCount(ctx)
		if err != nil {
			fmt.Println("Error getting user count:", err)
			continue
		}
		currentServer := ServerFilter{Hostname: server.Hostname, Users: userCount}

		if c.QueryParam("server-filter-"+server.Hostname) != "" {
			currentServer.FilteredFor = true
			selectedServers = append(selectedServers, server)
			isFilteredRequest = true
		}
		showMeetingsServerFiltered = append(showMeetingsServerFiltered, currentServer)
	}

	// update server stats percentages
	if serverStats.TotalMeetings > 0 {
		for iterator, count := range serverStats.ServerCounts {
			serverStats.ServerCounts[iterator].Percentage = float32(count.Meetings/serverStats.TotalMeetings) * 100
		}
	}

	if startDateParam != "" || endDateParam != "" || len(selectedServers) > 0 {
		isFilteredRequest = true
	}

	startDate, err = time.Parse("2006-01-02", startDateParam)
	endDate, err = time.Parse("2006-01-02", endDateParam)

	conn, err := pgx.Connect(ctx, confGet("DB_CONNECTION_STRING"))
	if err != nil {
		return err
	}
	dbQueries := db.New(conn)

	defer conn.Close(ctx)
	var meetings []MeetingListMeetingWrapper
	if isFilteredRequest {
		filteredMeetings, err := HandleFilteredRequest(ctx, dbQueries, startDate, endDate, selectedServers, &meetings)
		if err != nil {
			fmt.Println("error handeling filtered request:", err)
			return err
		}
		meetings = *filteredMeetings
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
			meetings = append(meetings, MeetingListMeetingWrapper{BBBEventsMeeting: rm, BbbHostname: rm.BbbHostname, Active: !m.MeetingEnded.Valid})
			// fmt.Println("DEBUG: m.MeetingEnded valid =", m.MeetingEnded.Valid, "m.MeetingEnded.time=", m.MeetingEnded.Time.Format(time.RFC3339), "time.Now() =", time.Now(), "m.MeetingEnded.Time.Before(time.Now()) =", m.MeetingEnded.Time.Before(time.Now()))
		}
	}

	for iterator, meeting := range meetings {
		if meeting.Active {
			tmpInt, err := dbQueries.GetActiveUserCountInMeetingByID(ctx, meeting.BBBEventsMeeting.InternalMeetingID)
			if err != nil {
				fmt.Println("Error getting active user count:", err)
			}
			meetings[iterator].UserCount = int(tmpInt)
		} else {
			usrCount, err := dbQueries.GetUserCountInMeetingByInternalID(ctx, meeting.BBBEventsMeeting.InternalMeetingID)
			if err != nil {
				fmt.Println("Error getting user count in meetings", err.Error())
				continue
			}
			meetings[iterator].UserCount = int(usrCount)
		}
	}

	sort.Slice(meetings, func(i, j int) bool {
		return meetings[i].BBBEventsMeeting.CreateTime.After(meetings[j].BBBEventsMeeting.CreateTime)
	})
	err = c.Render(http.StatusOK, "meetings", map[string]interface{}{"Request": struct {
		StartDate    string
		EndDate      string
		ServerFilter []ServerFilter
	}{StartDate: startDateParam, EndDate: endDateParam, ServerFilter: showMeetingsServerFiltered}, "Meetings": meetings, "ServerStats": serverStats})
	if err != nil {
		fmt.Println(err)
	}
	return nil
}

func HandleFilteredRequest(ctx context.Context, dbQueries *db.Queries, startDate time.Time, endDate time.Time, filteredServers []bbbServer, meetings *[]MeetingListMeetingWrapper) (filteredMeetings *[]MeetingListMeetingWrapper, err error) {

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
		return meetings, err
	}

	for _, m := range dbMeetings {
		rm := BBBEvents.ConvertDBToBBBMeeting(m)
		*meetings = append(*meetings, MeetingListMeetingWrapper{BBBEventsMeeting: rm, BbbHostname: rm.BbbHostname, Active: !m.MeetingEnded.Valid})
	}
	var meetingClone = slices.Clone(*meetings)
	if len(filteredServers) > 0 {
		var filteredSlice []MeetingListMeetingWrapper
		var allowedHostnames []string
		for _, server := range filteredServers {
			allowedHostnames = append(allowedHostnames, server.Hostname)
		}

		for _, meeting := range meetingClone {
			if slices.Contains(allowedHostnames, meeting.BbbHostname) {
				filteredSlice = append(filteredSlice, meeting)
			}
			continue
		}
		return &filteredSlice, nil

	}

	return &meetingClone, nil
}
