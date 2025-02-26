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
	"context"
	"fmt"
	"github.com/jackc/pgx/v5"
	"time"
)

func generateStatsByScope(ctx context.Context, scope string) (stats statistics, err error) {
	stats.ConferenceCount = make(map[string]int)
	stats.MaxUserCount = make(map[string]int)
	stats.ConferenceUsageHours = make(map[string]int)
	stats.ConferenceUsersUsageHours = make(map[string]int)

	conn, err := pgx.Connect(ctx, confGet("DB_CONNECTION_STRING"))
	if err != nil {
		return stats, fmt.Errorf("statsPage -> generateStatsByScope pgx.Connect: %w", err)
	}
	timeFrames := map[string]timeFrame{}
	if scope == "week" {
		timeFrames = generateWeekTimeFrames(time.Now())
	}
	if scope == "month" {
		timeFrames = generateMonthTimeFrames(time.Now())
	}
	if scope == "year" {
		timeFrames = generateYearTimeFrames(time.Now())
	}
	for index, tf := range timeFrames {
		var conferences int
		var userCount int
		var conferenceUsageHours time.Duration
		var userUsageHours time.Duration
		err = conn.QueryRow(ctx, "SELECT COUNT(*) FROM meetings WHERE create_time BETWEEN $1 AND $2", tf.start, tf.end).Scan(&conferences)
		if err != nil {
			return stats, fmt.Errorf("statsPage -> generateStatsByScope queryRow (meetings count): %w", err)
		}
		err = conn.QueryRow(ctx, "SELECT COUNT(*) FROM meetings WHERE create_time BETWEEN $1 AND $2", tf.start, tf.end).Scan(&userCount)
		if err != nil {
			return stats, fmt.Errorf("statsPage -> generateStatsByScope queryRow (user count): %w", err)
		}
		rows, err := conn.Query(ctx, "SELECT create_time, meeting_ended FROM meetings WHERE meeting_ended IS NOT NULL AND create_time BETWEEN $1 AND $2", tf.start, tf.end)
		if err != nil {
			return stats, fmt.Errorf("statsPage -> generateStatsByScope queryRow (meeting time): %w", err)
		}

		for rows.Next() {
			startTime, endTime := time.Time{}, time.Time{}
			err = rows.Scan(&startTime, &endTime)
			if err != nil {
				return stats, fmt.Errorf("statsPage -> generateStatsByScope (conference usage hours) rows.Scan: %w", err)
			}
			conferenceUsageHours += endTime.Sub(startTime)
		}
		rows.Close()

		rows, err = conn.Query(ctx, "SELECT join_timestamp, leave_timestamp FROM users WHERE users.leave_timestamp IS NOT NULL AND join_timestamp BETWEEN $1 AND $2", tf.start, tf.end)
		if err != nil {
			return stats, fmt.Errorf("statsPage -> generateStatsByScope queryRow (users usage hours time): %w", err)
		}

		for rows.Next() {
			startTime, endTime := time.Time{}, time.Time{}
			err = rows.Scan(&startTime, &endTime)
			if err != nil {
				return stats, fmt.Errorf("statsPage -> generateStatsByScope (users usage hours) rows.Scan: %w", err)
			}
			userUsageHours += endTime.Sub(startTime)
		}

		stats.ConferenceCount[index] = conferences
		stats.MaxUserCount[index] = userCount
		stats.ConferenceUsageHours[index] = int(conferenceUsageHours.Hours())
		stats.ConferenceUsersUsageHours[index] = int(userUsageHours.Hours())
		// used for normalizing the graph
		stats.HighestConferenceCount = findMaxInMap(stats.ConferenceCount)
		stats.HighestMaxUserCount = findMaxInMap(stats.MaxUserCount)
		stats.HighestConferenceUsageHours = findMaxInMap(stats.ConferenceUsageHours)

	}

	return stats, err
}
