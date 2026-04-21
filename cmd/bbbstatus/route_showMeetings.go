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
	db "bbbstatus/internal/database"
	"bbbstatus/locales"
	"bbbstatus/pkg/BBBAPI"
	"bbbstatus/pkg/BBBEvents"
	"context"
	"errors"
	"fmt"
	"net/http"
	"os"
	"slices"
	"sort"
	"strings"
	"sync"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgtype"
	"github.com/labstack/echo/v4"
	"github.com/nicksnyder/go-i18n/v2/i18n"
	"golang.org/x/text/language"
)

func init() { // Add all messages that are related to this file into the localization bundle
	var msgs = []i18n.Message{
		{ID: "meetingListHeader", Other: "meeting list"},
		{ID: "Meetings", Other: "meetings"},
		{ID: "Recordings", Other: "recordings"},
		{ID: "Statistics", Other: "statistics"},
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
		{ID: "MeetingRecordingEnabled", Other: "recordings are possible"},
		{ID: "MeetingRecordingDisabled", Other: "recordings are disabled"},
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
		{ID: "FleetStatus", Other: "Server Fleet Status"},
		{ID: "LiveMonitorandManagement", Other: "Live Monitor & Management"},
		{ID: "MeetingsSearchPlaceholder", Other: "Search for meeting / hostname..."},
		{ID: "MeetingsTableFirstColumnMeetingNameAndID", Other: "Meeting & ID"},
		{ID: "MeetingsTableSecondColumnServerHostname", Other: "Server Hostname"},
		{ID: "MeetingsTableThrirdColumnUserCount", Other: "User"},
		{ID: "MeetingsTableFourhtColumnActions", Other: "Actions"},
		{ID: "MeetingsPageMettingsDetailOngoingMeeting", Other: "Ongoing"},
		{ID: "MeetingsPageViewRecording", Other: "View recording"},
		{ID: "MeetingsMeetingDetailsFirstUserJoined", Other: "First User"},
		{ID: "MeetingsPageSearchPlaceholder", Other: "Search..."},
		{ID: "MeetingsPageDetailsPopupRecordingAvailableStateLabel", Other: "Available"},
		{ID: "MeetingsPageDetailsPopupRecordingNotAvailableStateLabel", Other: "Not available"},
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
func showMeetings(c echo.Context) (err error) {
	var cc = c.(*config.CustomContext)

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
	var startDate, endDate time.Time
	var requestLanguage = c.Request().Header.Get("Accept-Language")
	var ctx = context.WithValue(c.Request().Context(), locales.Translator("Translator"), c.Get("Translator"))
	locales.Localizer = i18n.NewLocalizer(locales.Bundle, requestLanguage, language.English.String())

	// handle filtered request
	startDateParam := c.FormValue("start-date")
	endDateParam := c.FormValue("end-date")
	searchQuery := c.FormValue("query")

	var selectedServers []config.BbbServer
	var showMeetingsServerFiltered []ServerFilter
	var serverMutex = new(sync.RWMutex)
	var wg = new(sync.WaitGroup)

	bigBlueButtonQueryStartTime := time.Now()
	bigBlueButtonServers := cc.Config.FindBBBServers("")
	wg.Add(len(bigBlueButtonServers))
	for _, server := range bigBlueButtonServers {
		go func(ctx context.Context) {
			defer wg.Done()
			var concurrentError error
			BbbApi := BBBAPI.API{Hostname: server.Hostname, Port: server.ApiPort, SharedSecret: server.SharedSecret, Timeout: new(time.Duration(server.APITimeout) * time.Second)}

			// fill server usage stats by meeting count
			var serverMeetings []BBBAPI.GetMeetingInfoResponse
			serverMeetings, concurrentError = BbbApi.GetMeetings(ctx)
			if concurrentError != nil {
				fmt.Printf("Error getting meetings: %v\n", concurrentError)
				return
			}
			serverMutex.Lock()
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
			serverMutex.Unlock()

			var userCount int
			userCount, concurrentError = BbbApi.GetServerUserCount(ctx)
			if concurrentError != nil {
				fmt.Println("Error getting user count:", concurrentError)
				return
			}
			currentServer := ServerFilter{Hostname: server.Hostname, Users: userCount}

			if c.QueryParam("server-filter-"+server.Hostname) != "" {
				currentServer.FilteredFor = true
				serverMutex.Lock()
				selectedServers = append(selectedServers, server)
				serverMutex.Unlock()
			}
			serverMutex.Lock()
			showMeetingsServerFiltered = append(showMeetingsServerFiltered, currentServer)
			serverMutex.Unlock()
		}(context.WithoutCancel(ctx))

	}
	wg.Wait()
	if time.Since(bigBlueButtonQueryStartTime).Seconds() > 5 {
		fmt.Println("# (WARNING) The big blue button servers are slow to query and took " + time.Since(bigBlueButtonQueryStartTime).String())
	}
	// update server stats percentages
	if serverStats.TotalMeetings > 0 {
		for iterator, count := range serverStats.ServerCounts {
			serverStats.ServerCounts[iterator].Percentage = float32(count.Meetings/serverStats.TotalMeetings) * 100
		}
	}

	startDate, err = time.Parse("2006-01-02", startDateParam)
	if err != nil && startDateParam != "" {
		fmt.Println("Error parsing start date:", err)
	}
	endDate, err = time.Parse("2006-01-02", endDateParam)
	if err != nil && startDateParam != "" {
		fmt.Println("Error parsing end date:", err)
	}
	if endDate.IsZero() {
		endDate = time.Now()
	}

	if startDate.IsZero() {
		startDate = endDate.AddDate(0, 0, -7)
	}

	conn, err := pgx.Connect(ctx, cc.Config.DatabaseConfig.DatabaseConnectionString)
	if err != nil {
		return err
	}
	dbQueries := db.New(conn)

	defer conn.Close(ctx)
	var meetings []MeetingListMeetingWrapper

	filteredMeetings, err := HandleFilteredRequest(ctx, dbQueries, startDate, endDate, selectedServers, searchQuery, &meetings)
	if err != nil {
		fmt.Println("error handling filtered request:", err)
		return err
	}
	meetings = *filteredMeetings

	userCountAggregationStartTime := time.Now()
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
	if time.Since(userCountAggregationStartTime).Seconds() > 5 {
		fmt.Println("# (WARNING) The GetActiveUserCountInMeetingByID and or the GetUserCountInMeetingByInternalID query returned slowly, some meetings may have been closed incorrectly or not at all; Aggregation time was " + time.Since(userCountAggregationStartTime).String())
	}

	sort.Slice(meetings, func(i, j int) bool {
		return meetings[i].BBBEventsMeeting.CreateTime.After(meetings[j].BBBEventsMeeting.CreateTime)
	})
	err = c.Render(http.StatusOK, "meetings", map[string]interface{}{"Request": struct {
		StartDate    string
		EndDate      string
		ServerFilter []ServerFilter
		Query        string
	}{StartDate: startDate.Format("2006-01-02"), EndDate: endDate.Format("2006-01-02"), ServerFilter: showMeetingsServerFiltered, Query: searchQuery}, "Meetings": meetings, "ServerStats": serverStats})
	if err != nil {
		fmt.Println(err)
	}
	return nil
}

func HandleFilteredRequest(ctx context.Context, dbQueries *db.Queries, startDate time.Time, endDate time.Time, filteredServers []config.BbbServer, Query string, meetings *[]MeetingListMeetingWrapper) (filteredMeetings *[]MeetingListMeetingWrapper, err error) {

	if endDate.IsZero() {
		endDate = time.Now()
	}
	endDate = endDate.AddDate(0, 0, 1).Add(-time.Second * 1) // make the end date inclusive.
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
		if Query == "" || strings.Contains(rm.BbbHostname, Query) || strings.Contains(rm.Name, Query) || strings.Contains(rm.InternalMeetingID, Query) {
			*meetings = append(*meetings, MeetingListMeetingWrapper{BBBEventsMeeting: rm, BbbHostname: rm.BbbHostname, Active: !m.MeetingEnded.Valid})
		}
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
			if strings.Contains(meeting.BbbHostname, Query) {
				// the user searched for this hostname
				filteredSlice = append(filteredSlice, meeting)
			}
			continue
		}
		return &filteredSlice, nil
	}

	return &meetingClone, nil
}
