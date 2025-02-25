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
	"math/rand/v2"
	"net/http"
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
		{ID: "StatsPageStatisticsOverviewHeader", Other: "Statistics overview"},
		{ID: "StatsPageStatisticsHeader", Other: "Statistics"},
		{ID: "StatsPageConferenceCountHeader", Other: "Conference count"},
		{ID: "StatsPageConferencesTableHeader", Other: "Conferences"},
		{ID: "StatsPageConferencesInToolTip", Other: "Conferences in"},
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

func statsPage(c echo.Context) error {
	scope := c.QueryParam("scope")

	templateData := templateStruct{}

	yearStats := statistics{
		ConferenceCount:           make(map[string]int),
		MaxUserCount:              make(map[string]int),
		ConferenceUsageHours:      make(map[string]int),
		ConferenceUsersUsageHours: make(map[string]int),
	}
	monthStats := statistics{
		ConferenceCount:           make(map[string]int),
		MaxUserCount:              make(map[string]int),
		ConferenceUsageHours:      make(map[string]int),
		ConferenceUsersUsageHours: make(map[string]int),
	}
	weekStats := statistics{
		ConferenceCount:           make(map[string]int),
		MaxUserCount:              make(map[string]int),
		ConferenceUsageHours:      make(map[string]int),
		ConferenceUsersUsageHours: make(map[string]int),
	}
	yearSampleData := map[string][4]int{ // sample data representing a year of bbb meetings.
		"Jan": {4, 150, 40, 500},
		"Feb": {3, 130, 35, 450},
		"Mar": {5, 180, 50, 600},
		"Apr": {6, 170, 60, 700},
		"May": {4, 160, 45, 550},
		"Jun": {7, 190, 70, 750},
		"Jul": {3, 140, 30, 400},
		"Aug": {5, 160, 55, 650},
		"Oct": {4, 150, 45, 500},
		"Sep": {6, 180, 60, 720},
		"Nov": {2, 120, 25, 300},
		"Dec": {8, 200, 75, 800},
	}

	monthSampleData := make(map[string][4]int)
	for _, name := range getDaysBetween(time.Now().Add(-time.Hour*24*30), time.Now()) {
		monthSampleData[name] = [4]int{rand.IntN(20), rand.IntN(20), rand.IntN(20), rand.IntN(20)}
	}

	weekSampleData := map[string][4]int{
		"Mon": {150, 4, 500, 40},
		"Tue": {450, 3, 130, 35},
		"Wed": {600, 5, 180, 50},
		"Thu": {500, 6, 160, 45},
		"Fri": {500, 4, 160, 45},
		"Sat": {750, 7, 190, 70},
		"Sun": {30, 3, 140, 400},
	}

	applyStats(yearSampleData, &yearStats)
	applyStats(monthSampleData, &monthStats)
	applyStats(weekSampleData, &weekStats)

	if scope != "" {
		switch scope {
		case "week":
			templateData.WeekActive = true
			templateData.Statistics = weekStats
			templateData.StatsOrder = []string{"Mon", "Tue", "Wed", "Thu", "Fri", "Sat", "Sun"}
		case "month":
			templateData.MonthActive = true
			templateData.Statistics = monthStats
			templateData.StatsOrder = getDaysBetween(time.Now().Add(-time.Hour*24*30), time.Now())
		case "year":
			templateData.YearActive = true
			templateData.Statistics = yearStats
			templateData.StatsOrder = []string{"Jan", "Feb", "Mar", "Apr", "May", "Jun", "Jul", "Aug", "Oct", "Sep", "Nov", "Dec"}

		}
		templateData.CurrentScope = scope
	}

	return c.Render(http.StatusOK, "statistics", templateData)
}

func applyStats(yearSampleData map[string][4]int, yearStats *statistics) {
	// Variables to track the highest values
	highestConferenceCount := 0
	highestMaxUserCount := 0
	highestConferenceUsageHours := 0
	highestConferenceUsersUsageHours := 0

	for unit, values := range yearSampleData {
		yearStats.ConferenceCount[unit] = values[0]
		yearStats.MaxUserCount[unit] = values[1]
		yearStats.ConferenceUsageHours[unit] = values[2]
		yearStats.ConferenceUsersUsageHours[unit] = values[3]

		// Update the highest values
		if values[0] > highestConferenceCount {
			highestConferenceCount = values[0]
		}
		if values[1] > highestMaxUserCount {
			highestMaxUserCount = values[1]
		}
		if values[2] > highestConferenceUsageHours {
			highestConferenceUsageHours = values[2]
		}
		if values[3] > highestConferenceUsersUsageHours {
			highestConferenceUsersUsageHours = values[3]
		}
	}

	// Assign the highest values to the struct
	yearStats.HighestConferenceCount = highestConferenceCount
	yearStats.HighestMaxUserCount = highestMaxUserCount
	yearStats.HighestConferenceUsageHours = highestConferenceUsageHours
	yearStats.HighestConferenceUsersUsageHours = highestConferenceUsersUsageHours
}

func getDaysBetween(startDate, endDate time.Time) []string {
	var days []string

	// Ensure startDate is before endDate
	if startDate.After(endDate) {
		return days
	}

	// Iterate from startDate to endDate
	for d := startDate; d.Before(endDate) || d.Equal(endDate); d = d.AddDate(0, 0, 1) {
		// Format the day and month
		formattedDay := fmt.Sprintf("%d %s", d.Day(), d.Month().String()[:3])
		days = append(days, formattedDay)
	}

	return days
}
