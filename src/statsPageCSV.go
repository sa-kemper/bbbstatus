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
	"BbbStatus/internal/StatsAggregator"
	"fmt"
	"github.com/labstack/echo/v4"
	"net/http"
	"strconv"
)

func statsPageCSV(c echo.Context) (err error) {
	var result string
	var StatsOrder []string
	var fileDate string
	startTimeStr := c.QueryParam("startTime")
	scope := c.QueryParam("scope")
	if scope == "" {
		scope = "week"
	}

	startTime, err := parseTargetTime(scope, startTimeStr)
	if err != nil {
		_ = renderError(c, "Error occurred generating cvs")
		return err
	}

	stats, err := StatsAggregator.GenerateStatsByScope(c.Request().Context(), scope, confGet("DB_CONNECTION_STRING"), startTime)
	if err != nil {
		_ = renderError(c, "Error occurred generating cvs")
		return err
	}
	result += Translate("StatsPageCSVTime") + ","
	result += Translate("StatsPageConferenceCountHeader") + ","
	result += Translate("StatsPageConferenceAttendeeCountHeader") + ","
	result += Translate("StatsPageTotalMeetingHoursTableHeader") + ","
	result += Translate("StatsPageConferenceUserHoursHeader") + ",\n"

	switch scope {
	case "week":
		StatsOrder = []string{"Mon", "Tue", "Wed", "Thu", "Fri", "Sat", "Sun"}[:len(stats.ConferenceCount)]
		_, week := startTime.ISOWeek()
		fileDate = fmt.Sprintf("%d-callendar-week-%d", startTime.Year(), week)
	case "month":
		_, currentWeek := startTime.ISOWeek()
		StatsOrder = []string{
			fmt.Sprintf("CW%d", currentWeek-4),
			fmt.Sprintf("CW%d", currentWeek-3),
			fmt.Sprintf("CW%d", currentWeek-2),
			fmt.Sprintf("CW%d", currentWeek-1),
		}[:len(stats.ConferenceCount)]
		fileDate = strconv.Itoa(startTime.Year()) + "-" + startTime.Month().String()
	case "year":
		StatsOrder = []string{"Jan", "Feb", "Mar", "Apr", "May", "Jun", "Jul", "Aug", "Oct", "Sep", "Nov", "Dec"}[:len(stats.ConferenceCount)]
		fileDate = strconv.Itoa(startTime.Year())
	}

	for _, index := range StatsOrder {
		result += Translate(index) + ","
		result += strconv.Itoa(stats.ConferenceCount[index]) + ","
		result += strconv.Itoa(stats.MaxUserCount[index]) + ","
		result += strconv.Itoa(stats.ConferenceUsageHours[index]) + ","
		result += strconv.Itoa(stats.ConferenceUsersUsageHours[index]) + ",\n"
	}

	c.Response().Header().Set("Content-Disposition", "attachment; filename="+fmt.Sprintf("bbbstatus-%s-statistics-%s.csv", scope, fileDate))
	return c.Blob(http.StatusOK, "text/csv", []byte(result))
}
