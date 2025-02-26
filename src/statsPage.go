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
	"BbbStatus/internal/BBBEvents"
	"fmt"
	"github.com/labstack/echo/v4"
	"github.com/nicksnyder/go-i18n/v2/i18n"
	"net/http"
	"strconv"
	"time"
)

func init() { // Add all messages that are related to this file into the localization bundle
	var msgs = []i18n.Message{
		{ID: "Week", Other: "Week"},
		{ID: "Month", Other: "Month"},
		{ID: "Year", Other: "Year"},
		{ID: "Mon", Other: "Mon"},
		{ID: "Tue", Other: "Tue"},
		{ID: "Wed", Other: "Wed"},
		{ID: "Thu", Other: "Thu"},
		{ID: "Fri", Other: "Fri"},
		{ID: "Sat", Other: "Sat"},
		{ID: "Sun", Other: "Sun"},
		{ID: "Jan", Other: "Jan"},
		{ID: "Feb", Other: "Feb"},
		{ID: "Mar", Other: "Mar"},
		{ID: "Apr", Other: "Apr"},
		{ID: "May", Other: "May"},
		{ID: "Jun", Other: "Jun"},
		{ID: "Jul", Other: "Jul"},
		{ID: "Aug", Other: "Aug"},
		{ID: "Sep", Other: "Sep"},
		{ID: "Oct", Other: "Oct"},
		{ID: "Nov", Other: "Nov"},
		{ID: "Dec", Other: "Dec"},
		{ID: "CW", Other: "CW"}, //Calendar week
		{ID: "StatsPageStatisticsOverviewHeader", Other: "Statistics overview"},
		{ID: "StatsPageStatisticsHeader", Other: "Statistics"},
		{ID: "StatsPageConferenceCountHeader", Other: "Conference count"},
		{ID: "StatsPageConferencesTableHeader", Other: "Conferences"},
		{ID: "StatsPageConferencesInToolTip", Other: "Conferences in"},
		{ID: "StatsPageConferenceAttendeeCountHeader", Other: "Conference attendees"},
		{ID: "StatsPageConferenceTotalMeetingHoursHeader", Other: "Conference total meeting hours"},
		{ID: "StatsPageAttendeeCountTableHeader", Other: "Attendees"},
		{ID: "StatsPageAttendeesInToolTip", Other: "Attendees "},
		{ID: "StatsPageTotalMeetingHoursTableHeader", Other: "Total meeting hours"},
		{ID: "StatsPageConferenceUsageHoursToolTip", Other: "Total meeting hours in"},
		{ID: "StatsPageConferenceUserHoursHeader", Other: "Total user hours"},
		{ID: "StatsPageConferencesUserHoursTableHeader", Other: "Total user hours"},
		{ID: "StatsPageConferenceUserHoursToolTip", Other: "Total user hours in"},
	}
	FrontendTextMessages = append(FrontendTextMessages, msgs...)
	for _, m := range BBBEvents.UserEventTextRepresentation { // Add user events text representation to the language strings.
		FrontendTextMessages = append(FrontendTextMessages, m)
	}
}

type statistics struct {
	ConferenceCount                  map[string]int
	MaxUserCount                     map[string]int
	ConferenceUsageHours             map[string]int
	ConferenceUsersUsageHours        map[string]int
	HighestConferenceCount           int
	HighestMaxUserCount              int
	HighestConferenceUsageHours      int
	HighestConferenceUsersUsageHours int
}

type templateStruct struct {
	CurrentScope string
	WeekActive   bool
	MonthActive  bool
	YearActive   bool
	Statistics   statistics
	StatsOrder   []string
}

type timeFrame struct {
	start, end time.Time
}

// findMaxInMap loops through a map[string]int object, ignores all indexes, and returns the highest value.
func findMaxInMap(in map[string]int) (max int) {
	for _, value := range in {
		if value > max {
			max = value
		}
	}
	return max
}

func statsPage(c echo.Context) error {
	scope := c.QueryParam("scope")

	templateData := templateStruct{}
	dbStats, err := generateStatsByScope(c.Request().Context(), scope)
	if err != nil {
		fmt.Println("A error occured while generating stats page", err)
		return c.Render(http.StatusInternalServerError, "errorPage", frontendError{ErrorTitle: "Internal Server Error", ErrorParagraph: "Error generating stats page"})
	}

	if scope != "" {
		switch scope {
		case "week":
			templateData.WeekActive = true
			templateData.Statistics = dbStats
			templateData.StatsOrder = []string{"Mon", "Tue", "Wed", "Thu", "Fri", "Sat", "Sun"}[:len(dbStats.ConferenceCount)]
		case "month":
			templateData.MonthActive = true
			templateData.Statistics = dbStats
			_, currentWeek := time.Now().ISOWeek()
			templateData.StatsOrder = []string{"CW" + strconv.Itoa(currentWeek-4), "CW" + strconv.Itoa(currentWeek-3), "CW" + strconv.Itoa(currentWeek-2), "CW" + strconv.Itoa(currentWeek-1)}[:len(dbStats.ConferenceCount)]
		case "year":
			templateData.YearActive = true
			templateData.Statistics = dbStats
			templateData.StatsOrder = []string{"Jan", "Feb", "Mar", "Apr", "May", "Jun", "Jul", "Aug", "Oct", "Sep", "Nov", "Dec"}[:len(dbStats.ConferenceCount)]

		}
		templateData.CurrentScope = scope
	}

	return c.Render(http.StatusOK, "statistics", templateData)
}
