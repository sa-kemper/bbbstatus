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
	"time"
)

func generateWeekTimeFrames(startTime time.Time) map[string]TimeFrame {
	wtf := make(map[string]TimeFrame)
	targetTime := startTime.AddDate(0, 0, 7)
	_, targetWeek := startTime.ISOWeek()
	if targetTime.After(time.Now()) {
		targetTime = time.Now()
	}
	start := startTime
	start = start.Add(-time.Hour * time.Duration(start.Hour()))
	start = start.Add(-time.Minute * time.Duration(start.Minute()))
	start = start.Add(-time.Second * time.Duration(start.Second()))
	var firstRunFlag = true
	for {
		if firstRunFlag && start.Weekday() != time.Monday {
			start = start.AddDate(0, 0, -1)
			continue
		}
		firstRunFlag = false
		_, currentWeek := start.ISOWeek()
		if currentWeek > targetWeek || len(wtf) == 7 || start.After(time.Now()) || start.After(targetTime) {
			break
		}
		wtf[start.Weekday().String()[:3]] = TimeFrame{
			start: start,
			end:   start.AddDate(0, 0, 1),
		}
		start = start.AddDate(0, 0, 1)

	}

	return wtf
}
