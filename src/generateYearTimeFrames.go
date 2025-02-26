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

import "time"

func generateYearTimeFrames(start time.Time) map[string]timeFrame {
	// get to the start of the month by striping the day, hour, minute and second information.
	start = start.AddDate(0, 0, -start.Day()+1)
	start = start.Add(-time.Hour * time.Duration(start.Hour()))
	start = start.Add(-time.Minute * time.Duration(start.Minute()))
	start = start.Add(-time.Second * time.Duration(start.Second()))

	ytf := make(map[string]timeFrame)
	var firstRunFlag = true
	start = start.AddDate(-1, 0, 0)
	for {
		if firstRunFlag && start.Month() != time.January {
			start = start.AddDate(0, 1, 0)
			continue
		}
		if start.Month() == time.January {
			firstRunFlag = false
		}
		if start.After(time.Now()) {
			break
		}
		ytf[start.Month().String()[:3]] = timeFrame{start: start, end: start.AddDate(0, 1, 0)}
		start = start.AddDate(0, 1, 0)
	}

	return ytf
}
