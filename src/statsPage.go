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
	"BbbStatus/internal/StatsAggregator"
	"fmt"
	"github.com/labstack/echo/v4"
	"github.com/nicksnyder/go-i18n/v2/i18n"
	"net/http"
	"strconv"
	"strings"
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
		{ID: "StatsPageSelectScopeTime", Other: "Select start time"},
	}
	FrontendTextMessages = append(FrontendTextMessages, msgs...)
	for _, m := range BBBEvents.UserEventTextRepresentation { // Add user events text representation to the language strings.
		FrontendTextMessages = append(FrontendTextMessages, m)
	}
}

type templateStruct struct {
	CurrentScope string
	StartTime    string

	StartTimeMin string
	StartTimeMax string

	WeekActive  bool
	MonthActive bool
	YearActive  bool
	Statistics  StatsAggregator.Statistics
	StatsOrder  []string
}

func statsPage(c echo.Context) (err error) {
	scope := c.QueryParam("scope")
	startTime := c.QueryParam("startTime")
	if scope == "" {
		scope = "week"
	}

	var targetTime time.Time
	if startTime != "" {
		if scope == "week" {
			var yearStr, weekStr string
			if strings.Contains(startTime, "-") {
				yearStr = strings.Split(startTime, "-")[0]
				weekStr = strings.Split(startTime, "-")[1]
				weekStr = strings.TrimLeft(weekStr, "W")
				weekStr = strings.TrimLeft(weekStr, "0")
			}
			year, err := strconv.Atoi(yearStr)
			if err != nil {
				return err
			}

			targetTime = time.Date(year, 1, 1, 1, 0, 0, 0, time.UTC)
			for {
				targetTime = targetTime.AddDate(0, 0, 1)
				_, week := targetTime.ISOWeek()
				if strconv.Itoa(week) == weekStr {
					break
				}
			}
		}
		if scope == "month" {
			targetTime, err = time.Parse("2006-01", startTime)
			if err != nil {
				return err
			}
			targetTime = targetTime.AddDate(0, 0, 1)
		}
		if scope == "year" {
			startTime = strings.Split(startTime, "-")[0]
			targetTime, err = time.Parse("2006", startTime)
			if err != nil {
				return err
			}
			targetTime = targetTime.AddDate(0, 1, 1)
		}
	} else {
		targetTime = time.Now()
	}

	templateData := templateStruct{}
	dbStats, err := StatsAggregator.GenerateStatsByScope(c.Request().Context(), scope, confGet("DB_CONNECTION_STRING"), targetTime)
	if err != nil {
		fmt.Println("A error occured while generating stats page", err)
		return c.Render(http.StatusInternalServerError, "errorPage", frontendError{ErrorTitle: "Internal Server Error", ErrorParagraph: "Error generating stats page"})
	}

	earliestDataTimestamp, err := StatsAggregator.GetEarliestStatDate(c.Request().Context(), confGet("DB_CONNECTION_STRING"))
	if err != nil {
		fmt.Println("A error occured while generating stats page (get earliest data point)", err)
		return c.Render(http.StatusInternalServerError, "errorPage", frontendError{ErrorTitle: "Internal Server Error", ErrorParagraph: "Error generating stats page"})
	}

	if startTime != "" {
		templateData.StartTime = startTime
	}

	switch scope {
	case "week":
		minYear, minWeek := earliestDataTimestamp.ISOWeek()
		maxYear, maxWeek := time.Now().ISOWeek()

		templateData.WeekActive = true
		templateData.Statistics = dbStats
		templateData.StatsOrder = []string{"Mon", "Tue", "Wed", "Thu", "Fri", "Sat", "Sun"}[:len(dbStats.ConferenceCount)]
		templateData.StartTimeMin = fmt.Sprintf("%v-W%v", minYear, minWeek)
		templateData.StartTimeMax = fmt.Sprintf("%v-W%v", maxYear, maxWeek)
	case "month":
		templateData.MonthActive = true
		templateData.Statistics = dbStats
		_, currentWeek := time.Now().ISOWeek()
		templateData.StatsOrder = []string{"CW" + strconv.Itoa(currentWeek-4), "CW" + strconv.Itoa(currentWeek-3), "CW" + strconv.Itoa(currentWeek-2), "CW" + strconv.Itoa(currentWeek-1)}[:len(dbStats.ConferenceCount)]
		templateData.StartTimeMin = fmt.Sprintf("%v-%d", earliestDataTimestamp.Year(), earliestDataTimestamp.Month())
		templateData.StartTimeMax = fmt.Sprintf("%v-%d", time.Now().Year(), time.Now().Month())

	case "year":
		templateData.YearActive = true
		templateData.Statistics = dbStats
		templateData.StatsOrder = []string{"Jan", "Feb", "Mar", "Apr", "May", "Jun", "Jul", "Aug", "Oct", "Sep", "Nov", "Dec"}[:len(dbStats.ConferenceCount)]
		templateData.StartTimeMin = fmt.Sprintf("%v-%v-%v", earliestDataTimestamp.Year(), earliestDataTimestamp.Month(), earliestDataTimestamp.Day())
		templateData.StartTimeMax = fmt.Sprintf("%v-%v-%v", time.Now().Year(), time.Now().Month(), time.Now().Day())

	}
	templateData.CurrentScope = scope

	return c.Render(http.StatusOK, "statistics", templateData)
}
