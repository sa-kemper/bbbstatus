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
	"bbbstatus/internal/StatsAggregator"
	"fmt"
	"github.com/labstack/echo/v4"
	"net/http"
	"slices"
	"sort"
	"strconv"
	"strings"
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
		StatsOrder = generateMonthStatIndexes(stats.TimeFrames)[:len(stats.ConferenceCount)]
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

func generateMonthStatIndexes(timeFrames []StatsAggregator.TimeFrame) (mtfIndex []string) {
	for _, tf := range timeFrames {
		_, week := tf.Start.ISOWeek()
		if !slices.Contains(mtfIndex, fmt.Sprintf("CW%d", week)) {
			mtfIndex = append(mtfIndex, fmt.Sprintf("CW%d", week))
		}

		_, week = tf.End.ISOWeek()
		if !slices.Contains(mtfIndex, fmt.Sprintf("CW%d", week)) {
			mtfIndex = append(mtfIndex, fmt.Sprintf("CW%d", week))
		}
	}
	sort.Slice(mtfIndex, func(i, j int) bool {
		digitI, err := strconv.Atoi(strings.TrimLeft(mtfIndex[i], "CW"))
		if err != nil {
			fmt.Println("err", err)
		}
		digitJ, err := strconv.Atoi(strings.TrimLeft(mtfIndex[j], "CW"))
		if err != nil {
			fmt.Println("err", err)
		}

		return digitI < digitJ
	})
	return
}
