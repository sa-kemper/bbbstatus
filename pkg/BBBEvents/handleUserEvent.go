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

package BBBEvents

import (
	db "bbbstatus/internal/database"
	"context"
	"fmt"

	"github.com/jackc/pgx/v5/pgtype"
)

// handleUserEvent is a function to handle edge cases of the insert into 'user_events' table.
// some events do not provide a user or a
func handleUserEvent(ctx context.Context, dbQueries *db.Queries, user *User, meeting Meeting, b *BaseEvent) error {
	var userInRequestExists bool
	var err error
	EventID := BBBEventType(b.Data.ID)
	if EventID == "user-joined" {
		return handleUserJoined(ctx, dbQueries, user, meeting, b)
	}
	if EventID == "user-left" {
		err = dbQueries.LeaveUserByID(ctx, db.LeaveUserByIDParams{LeaveTimestamp: pgtype.Timestamp{Time: b.GetTimestamp(), Valid: true}, InternalUserID: user.InternalUserID})
		if err != nil {
			fmt.Printf("Data invalidation Error -> updating users.leave_timestamp (userID:%v): %v\n", user.InternalUserID, err)
			return err
		}
	}

	if user == nil {
		if EventID == EventMeetingScreenshareStopped || EventID == EventMeetingScreenshareStarted {
			user = &User{}
			dbUser, err := dbQueries.GetPresenterUserByMeetingID(ctx, meeting.InternalMeetingID) // huh?
			if err != nil {                                                                      // handle the failure of the query
				fmt.Println("Failed obtaining the presenters userID ->", err)
			}

			user.Name = dbUser.Name

		} else {
			return fmt.Errorf("user is unexpectedly nil")
		}
	}
	userInRequestExists, err = dbQueries.GetUserExistsByID(ctx, user.InternalUserID)
	if err != nil {
		fmt.Println(err)
		return err
	}
	if !userInRequestExists {
		return fmt.Errorf("user %s is not present. This is unexpected behaviour, check your DB integrity and the API authorisation", user.InternalUserID)
	}
	err = dbQueries.InsertUserEvent(ctx, db.InsertUserEventParams{
		EventTimestamp:    pgtype.Timestamp{Time: b.GetTimestamp(), Valid: true},
		InternalUserID:    user.InternalUserID,
		EventType:         string(EventID),
		InternalMeetingID: meeting.InternalMeetingID,
	})
	if err != nil {
		fmt.Println(err)
		return err
	}
	return nil
}
