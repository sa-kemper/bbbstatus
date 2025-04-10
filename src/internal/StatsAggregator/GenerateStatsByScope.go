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
	db "bbbstatus/internal/database"
	"context"
	"fmt"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgtype"
	"time"
)

func GenerateStatsByScope(ctx context.Context, scope string, dbConnectionString string, targetTime time.Time) (stats Statistics, err error) {
	stats.ConferenceCount = make(map[string]int)
	stats.MaxUserCount = make(map[string]int)
	stats.ConferenceUsageHours = make(map[string]int)
	stats.ConferenceUsersUsageHours = make(map[string]int)

	conn, err := pgx.Connect(ctx, dbConnectionString)
	if err != nil {
		return stats, fmt.Errorf("statsPage -> generateStatsByScope pgx.Connect: %w", err)
	}
	defer conn.Close(ctx)

	dbQueries := db.New(conn)

	timeFrames := map[string]TimeFrame{}
	if scope == "week" {
		timeFrames = generateWeekTimeFrames(targetTime)
	}
	if scope == "month" {
		timeFrames = generateMonthTimeFrames(targetTime)
	}
	if scope == "year" {
		timeFrames = generateYearTimeFrames(targetTime)
	}
	for index, tf := range timeFrames {
		var conferences int64
		var userCount int64
		var conferenceUsageHours time.Duration
		var userUsageHours time.Duration
		conferences, err = dbQueries.GetMeetingCountBetweenDates(
			ctx,
			db.GetMeetingCountBetweenDatesParams{
				CreateTime:   pgtype.Timestamp{Valid: true, Time: tf.Start},
				CreateTime_2: pgtype.Timestamp{Valid: true, Time: tf.End},
			},
		)
		if err != nil {
			return stats, fmt.Errorf("statsPage -> generateStatsByScope queryRow (meetings count): %w", err)
		}
		userCount, err = dbQueries.GetUserCountBetweenDates(ctx, db.GetUserCountBetweenDatesParams{
			JoinTimestamp:   pgtype.Timestamp{Time: tf.Start, Valid: true},
			JoinTimestamp_2: pgtype.Timestamp{Time: tf.End, Valid: true},
		})
		if err != nil {
			return stats, fmt.Errorf("statsPage -> generateStatsByScope queryRow (user count): %w", err)
		}

		meetings, err := dbQueries.GetMeetingsBetweenDates(ctx, db.GetMeetingsBetweenDatesParams{
			CreateTime:   pgtype.Timestamp{Valid: true, Time: tf.Start},
			CreateTime_2: pgtype.Timestamp{Valid: true, Time: tf.End},
		})
		if err != nil {
			return stats, fmt.Errorf("statsPage -> generateStatsByScope queryRow (meeting time): %w", err)
		}

		for _, meeting := range meetings {
			startTime, endTime := meeting.CreateTime.Time, meeting.MeetingEnded.Time
			if !meeting.MeetingEnded.Valid {
				continue
			}
			if meeting.MeetingEnded.Time.IsZero() {
				continue
			}
			if meeting.CreateTime.Time.Before(tf.Start) == true || meeting.MeetingEnded.Time.After(tf.End) == true {
				continue
			}

			conferenceUsageHours += endTime.Sub(startTime)
		}

		users, err := dbQueries.GetUsersWhoJoinedBetween(ctx, db.GetUsersWhoJoinedBetweenParams{
			JoinTimestamp:   pgtype.Timestamp{Valid: true, Time: tf.Start},
			JoinTimestamp_2: pgtype.Timestamp{Valid: true, Time: tf.End},
		})
		if err != nil {
			return stats, fmt.Errorf("statsPage -> generateStatsByScope queryRow (users usage hours time): %w", err)
		}
		for _, user := range users {
			if !user.LeaveTimestamp.Valid {
				continue
			}
			if user.LeaveTimestamp.Time.IsZero() == true {
				continue
			}
			userUsageHours += user.LeaveTimestamp.Time.Sub(user.JoinTimestamp.Time)
		}

		stats.ConferenceCount[index] = int(conferences)
		stats.MaxUserCount[index] = int(userCount)
		stats.ConferenceUsageHours[index] = int(conferenceUsageHours.Hours())
		stats.ConferenceUsersUsageHours[index] = int(userUsageHours.Hours())
		// used for normalizing the graph
		stats.HighestConferenceCount = findMaxInMap(stats.ConferenceCount)
		stats.HighestMaxUserCount = findMaxInMap(stats.MaxUserCount)
		stats.HighestConferenceUsageHours = findMaxInMap(stats.ConferenceUsageHours)
		for _, tf := range timeFrames {
			stats.TimeFrames = append(stats.TimeFrames, tf)
		}
	}

	return stats, err
}
