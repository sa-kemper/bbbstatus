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
	"bbbstatus/internal/StatsAggregator"
	"database/sql"
	"errors"
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
		{ID: "StatsPageStatisticsOverviewHeader", Other: "statistics overview"},
		{ID: "StatsPageStatisticsHeader", Other: "statistics"},
		{ID: "StatsPageConferenceCountHeader", Other: "conference count"},
		{ID: "StatsPageConferencesTableHeader", Other: "conferences"},
		{ID: "StatsPageConferencesInToolTip", Other: "conferences in"},
		{ID: "StatsPageConferenceAttendeeCountHeader", Other: "conference attendees"},
		{ID: "StatsPageConferenceTotalMeetingHoursHeader", Other: "conference total meeting hours"},
		{ID: "StatsPageAttendeeCountTableHeader", Other: "attendees"},
		{ID: "StatsPageAttendeesInToolTip", Other: "attendees "},
		{ID: "StatsPageTotalMeetingHoursTableHeader", Other: "total meeting hours"},
		{ID: "StatsPageConferenceUsageHoursToolTip", Other: "total meeting hours in"},
		{ID: "StatsPageConferenceUserHoursHeader", Other: "total user hours"},
		{ID: "StatsPageConferencesUserHoursTableHeader", Other: "total user hours"},
		{ID: "StatsPageConferenceUserHoursToolTip", Other: "total user hours in"},
		{ID: "StatsPageSelectScopeTime", Other: "select start time"},
		{ID: "StatsPageSelectScopeTimePrintLabel", Other: "date\t"},
		{ID: "StatsPageCSVTime", Other: "time"},
		{ID: "InternalServerError", Other: "internal Server Error"},
		{ID: "NoFutureTimeAllowed", Other: "selected time point lays in the future"},
		{ID: "StatsPageGenError", Other: "error generating statistics"},
		{ID: "NoStatsAvailableYet", Other: "no statistics available yet"},
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
	if targetTime.After(time.Now()) {
		return renderError(c, "NoFutureTimeAllowed")
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
		return renderError(c, "StatsPageGenError")
	}

	// Get the earliest statistics date
	earliestDataTimestamp, err := StatsAggregator.GetEarliestStatDate(
		c.Request().Context(),
		confGet("DB_CONNECTION_STRING"),
	)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return renderError(c, "NoStatsAvailableYet")
		}

		fmt.Printf("Error getting earliest stat date: %v\n", err)
		return renderError(c, "StatsPageGenError")
	}

	// Set start time if provided
	if startTime != "" {
		templateData.StartTime = startTime
	}

	switch scope {
	case "week":
		minYear, minWeek := earliestDataTimestamp.ISOWeek()
		maxYear, maxWeek := targetTime.ISOWeek()
		currYear, currWeek := time.Now().ISOWeek()
		templateData.WeekActive = true
		templateData.Statistics = dbStats
		templateData.StatsOrder = []string{"Mon", "Tue", "Wed", "Thu", "Fri", "Sat", "Sun"}[:len(dbStats.ConferenceCount)]
		templateData.StartTimeMin = fmt.Sprintf("%d-W%d", minYear, minWeek)
		templateData.StartTimeMax = fmt.Sprintf("%d-W%d", maxYear, maxWeek)
		if templateData.StartTime == "" {
			templateData.StartTime = fmt.Sprintf("%d-W%02d", currYear, currWeek)
		}

	case "month":
		templateData.MonthActive = true
		templateData.Statistics = dbStats
		templateData.StatsOrder = generateMonthStatIndexes(dbStats.TimeFrames)[:len(dbStats.ConferenceCount)]
		templateData.StartTimeMin = fmt.Sprintf("%d-%d", earliestDataTimestamp.Year(), earliestDataTimestamp.Month())
		templateData.StartTimeMax = fmt.Sprintf("%d-%d", time.Now().Year(), time.Now().Month())
		if templateData.StartTime == "" {
			templateData.StartTime = fmt.Sprintf("%d-%02d", time.Now().Year(), time.Now().Month())
		}

	case "year":
		templateData.YearActive = true
		templateData.Statistics = dbStats
		templateData.StatsOrder = []string{"Jan", "Feb", "Mar", "Apr", "May", "Jun", "Jul", "Aug", "Oct", "Sep", "Nov", "Dec"}[:len(dbStats.ConferenceCount)]
		templateData.StartTimeMin = earliestDataTimestamp.Format("2006-01-02")
		templateData.StartTimeMax = time.Now().Format("2006-01-02")
		if templateData.StartTime == "" {
			templateData.StartTime = time.Now().Format("2006-01-02")
		}
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
			return targetTime, nil
		}
		targetTime = targetTime.AddDate(0, 0, 1)
	}
}

// renderError renders an error page with a standard message
func renderError(c echo.Context, message string) error {
	return c.Render(http.StatusInternalServerError, "errorPage", frontendError{
		ErrorTitle:     "InternalServerError",
		ErrorParagraph: message,
	})
}
