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

package StatsAggregator

import (
	"fmt"
	"time"
)

func generateMonthTimeFrames(start time.Time) map[string]TimeFrame {
	start = start.AddDate(0, 0, -start.Day()+1)
	start = start.Add(-time.Hour * time.Duration(start.Hour()))
	start = start.Add(-time.Minute * time.Duration(start.Minute()))
	start = start.Add(-time.Second * time.Duration(start.Second()))
	mtw := make(map[string]TimeFrame)
	for _, cw := range collectCalendarWeeksForMonth(start.Year(), start.Month()) {
		_, week := cw.Start.ISOWeek()
		mtw[fmt.Sprintf("CW%d", week)] = cw
	}
	return mtw
}

func generateCalendarWeeks(year int) (result map[int]TimeFrame) {
	start := time.Date(year, 1, -7, 0, 0, 0, 0, time.Local)
	result = make(map[int]TimeFrame)
	for {
		if start.Year() > year {
			break
		}
		if start.Weekday() == time.Monday {
			_, week := start.ISOWeek()
			if _, ok := result[week]; ok {
				return
			}
			result[week] = TimeFrame{start, start.AddDate(0, 0, 6)}
		}
		start = start.AddDate(0, 0, 1)
	}
	return
}

func collectCalendarWeeksForMonth(year int, month time.Month) (result []TimeFrame) {
	_, start := time.Date(year, month, 1, 0, 0, 1, 0, time.UTC).ISOWeek()
	_, end := time.Date(year, month, 1, 0, 0, 1, 0, time.UTC).AddDate(0, 1, 0).ISOWeek()
	if month == time.December {
		end = 52
	}
	if month == time.January {
		start = 1
	}
	calendarWeeks := generateCalendarWeeks(year)
	for i := start; i <= end; i++ {
		result = append(result, calendarWeeks[i])
	}
	return
}
