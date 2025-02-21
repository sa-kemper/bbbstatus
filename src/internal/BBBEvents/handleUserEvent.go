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
	"context"
	"fmt"
	"github.com/jackc/pgx/v5"
)

func handleUserEvent(ctx context.Context, conn *pgx.Conn, user *User, tx pgx.Tx, meeting Meeting, b *BaseEvent) error {
	var userInRequestExists bool
	var err error
	if b.Data.ID == "user-joined" {
		return handleUserJoined(ctx, conn, user, meeting, b)
	}
	if b.Data.ID == "user-left" {
		_, err = conn.Exec(ctx, "UPDATE users SET leave_timestamp = $1 WHERE internal_user_id = $2", b.GetTimestamp(), user.InternalUserID)
		if err != nil {
			fmt.Printf("Data invalidation Error -> updating users.leave_timestamp (userID:%v): %v\n", user.InternalUserID, err)
			return err
		}
	}

	if user == nil {
		if b.Data.ID == EventMeetingScreenshareStopped || b.Data.ID == EventMeetingScreenshareStarted {
			user = &User{}
			err = conn.QueryRow(ctx, "SELECT internal_user_id FROM user_events WHERE event_type = 'user-presenter-assigned' ORDER BY event_timestamp DESC LIMIT 1").Scan(&user.InternalUserID) // Obtain the current presenter's InternalUserId
			if err != nil {                                                                                                                                                                    // handle the failure of the query
				fmt.Println("Failed obtaining the presenters userID ->", err)
			} else { // obtain the username of the current presenter, if possible.
				err = conn.QueryRow(ctx, "SELECT name FROM users WHERE internal_user_id = $1", user.InternalUserID).Scan(&user.Name)
			}

		} else {
			return fmt.Errorf("user is unexpectedly nil")
		}
	}
	userInRequestExists, err = userExists(ctx, conn, user.InternalUserID)
	if err != nil {
		fmt.Println(err)
		return err
	}
	if !userInRequestExists {
		return fmt.Errorf("user %s is not present. This is unexpected behaviour, check your DB integrity and the API authorisation", user.InternalUserID)
	}
	_, err = tx.Exec(context.Background(), "INSERT INTO user_events (internal_meeting_id, internal_user_id, event_type, event_timestamp) VALUES ($1, $2, $3, $4)", meeting.InternalMeetingID, user.InternalUserID, b.Data.ID, b.GetTimestamp())
	if err != nil {
		fmt.Println(err)
		return err
	}
	err = tx.Commit(context.Background())
	if err != nil {
		fmt.Println(err)
		return err
	}
	return nil
}
