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
		{ID: "StatsPageSelectScopeTimePrintLabel", Other: "Date\t"},
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

// statsPage handles the rendering of the statistics page based on scope and time parameters
//
// Parameters:
//
//	c (echo.Context): The Echo context containing request information
//
// Returns:
//
//	error: Returns an error if stats generation or rendering fails, nil otherwise
func statsPage(c echo.Context) (err error) {
	// Get and set default scope
	scope := c.QueryParam("scope")
	if scope == "" {
		scope = "week"
	}

	// Parse start time parameter
	startTime := c.QueryParam("startTime")
	targetTime, err := parseTargetTime(scope, startTime)
	if err != nil {
		return err
	}

	// Initialize template data
	templateData := templateStruct{
		CurrentScope: scope,
	}

	// Generate statistics from database
	dbStats, err := StatsAggregator.GenerateStatsByScope(
		c.Request().Context(),
		scope,
		confGet("DB_CONNECTION_STRING"),
		targetTime,
	)
	if err != nil {
		fmt.Printf("Error generating stats page: %v\n", err)
		return renderError(c, "Error generating stats page")
	}

	// Get the earliest statistics date
	earliestDataTimestamp, err := StatsAggregator.GetEarliestStatDate(
		c.Request().Context(),
		confGet("DB_CONNECTION_STRING"),
	)
	if err != nil {
		fmt.Printf("Error getting earliest stat date: %v\n", err)
		return renderError(c, "Error generating stats page")
	}

	// Set start time if provided
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
		templateData.StartTimeMin = fmt.Sprintf("%d-W%d", minYear, minWeek)
		templateData.StartTimeMax = fmt.Sprintf("%d-W%d", maxYear, maxWeek)

	case "month":
		templateData.MonthActive = true
		templateData.Statistics = dbStats
		_, currentWeek := time.Now().ISOWeek()
		templateData.StatsOrder = []string{
			fmt.Sprintf("CW%d", currentWeek-4),
			fmt.Sprintf("CW%d", currentWeek-3),
			fmt.Sprintf("CW%d", currentWeek-2),
			fmt.Sprintf("CW%d", currentWeek-1),
		}[:len(dbStats.ConferenceCount)]
		templateData.StartTimeMin = fmt.Sprintf("%d-%d", earliestDataTimestamp.Year(), earliestDataTimestamp.Month())
		templateData.StartTimeMax = fmt.Sprintf("%d-%d", time.Now().Year(), time.Now().Month())

	case "year":
		templateData.YearActive = true
		templateData.Statistics = dbStats
		templateData.StatsOrder = []string{"Jan", "Feb", "Mar", "Apr", "May", "Jun", "Jul", "Aug", "Oct", "Sep", "Nov", "Dec"}[:len(dbStats.ConferenceCount)]
		templateData.StartTimeMin = earliestDataTimestamp.Format("2006-01-02")
		templateData.StartTimeMax = time.Now().Format("2006-01-02")
	}

	return c.Render(http.StatusOK, "statistics", templateData)
}

// parseTargetTime parses the target time based on scope and startTime string
//
// Parameters:
//
//	scope (string): The time scope (week, month, or year)
//	startTime (string): The starting time string
//
// Returns:
//
//	time.Time: The parsed target time
//	error: Error if parsing fails
func parseTargetTime(scope, startTime string) (time.Time, error) {
	if startTime == "" {
		return time.Now(), nil
	}

	switch scope {
	case "week":
		return parseWeekTime(startTime)
	case "month":
		t, err := time.Parse("2006-01", startTime)
		if err != nil {
			return time.Time{}, err
		}
		return t.AddDate(0, 0, 1), nil
	case "year":
		t, err := time.Parse("2006", strings.Split(startTime, "-")[0])
		if err != nil {
			return time.Time{}, err
		}
		return t.AddDate(0, 1, 1), nil
	default:
		return time.Now(), nil
	}
}

// parseWeekTime parses a week-based time string in format "YYYY-WWW"
func parseWeekTime(startTime string) (time.Time, error) {
	if !strings.Contains(startTime, "-") {
		return time.Time{}, fmt.Errorf("invalid week format")
	}

	parts := strings.Split(startTime, "-")
	year, err := strconv.Atoi(parts[0])
	if err != nil {
		return time.Time{}, err
	}

	weekStr := strings.TrimLeft(strings.TrimPrefix(parts[1], "W"), "0")
	week, err := strconv.Atoi(weekStr)
	if err != nil {
		return time.Time{}, err
	}

	targetTime := time.Date(year, 1, 1, 0, 0, 0, 0, time.UTC)
	for {
		_, currentWeek := targetTime.ISOWeek()
		if currentWeek == week {
			return targetTime.AddDate(0, 0, 7), nil
		}
		targetTime = targetTime.AddDate(0, 0, 1)
	}
}

// renderError renders an error page with a standard message
func renderError(c echo.Context, message string) error {
	return c.Render(http.StatusInternalServerError, "errorPage", frontendError{
		ErrorTitle:     "Internal Server Error",
		ErrorParagraph: message,
	})
}
