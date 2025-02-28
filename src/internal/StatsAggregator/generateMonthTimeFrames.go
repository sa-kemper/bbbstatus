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
	"strconv"
	"time"
)

func generateMonthTimeFrames(start time.Time) map[string]TimeFrame {
	//start = start.AddDate(0, -1, 0)
	targetTime := start.AddDate(0, 1, -2)
	if targetTime.After(time.Now()) {
		targetTime = time.Now()
	}
	start = start.AddDate(0, 0, -start.Day()+1)
	start = start.Add(-time.Hour * time.Duration(start.Hour()))
	start = start.Add(-time.Minute * time.Duration(start.Minute()))
	start = start.Add(-time.Second * time.Duration(start.Second()))

	mtf := make(map[string]TimeFrame)
	for {

		if start.Month() > targetTime.Month() || len(mtf) == 4 || start.After(targetTime) || start.After(time.Now()) {
			break
		}
		_, week := start.ISOWeek()
		mtf["CW"+strconv.Itoa(week)] = TimeFrame{start: start, end: start.AddDate(0, 0, 7)}
		start = start.AddDate(0, 0, 7)
	}
	return mtf
}
