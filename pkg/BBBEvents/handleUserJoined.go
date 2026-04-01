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
	"bbbstatus/locales"
	"bbbstatus/pkg/namegen"
	"context"
	"errors"
	"fmt"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgtype"
)

func handleUserJoined(ctx context.Context, dbQueries *db.Queries, user *User, meeting Meeting, b *BaseEvent) (err error) {
	var userInRequestExists bool

	if user == nil {
		return fmt.Errorf("user is unexpectedly nil")
	}

	userInRequestExists, err = dbQueries.GetUserExistsByID(ctx, user.InternalUserID)
	if err != nil {
		if !errors.Is(err, pgx.ErrNoRows) {
			fmt.Println("error occurred while checking if user exists: ", err)
			return err
		}
		userInRequestExists = false
	}
	//fmt.Println("DEBUG: handleUserJoined -> user exists: ", userInRequestExists)

	if !userInRequestExists {
		usersInTheCurrentMeeting, err := dbQueries.GetUsersInMeetingByID(ctx, meeting.InternalMeetingID)
		if err != nil {
			if !errors.Is(err, pgx.ErrNoRows) {
				fmt.Println("error occurred while checking for users in the meeting: ", meeting.InternalMeetingID, err)
			}
		}

		userNames := make([]string, len(usersInTheCurrentMeeting))
		for i := 0; i < len(usersInTheCurrentMeeting); i++ {
			userNames[i] = usersInTheCurrentMeeting[i].Name
		}

		GdprName := namegen.GenerateUnique(ctx.Value(locales.ServerLanguage("SERVER_LANG")).(string), &userNames)
		err = dbQueries.InsertUser(ctx, db.InsertUserParams{InternalUserID: user.InternalUserID, ExternalUserID: user.ExternalUserID, Name: user.Name, GdprName: GdprName, Role: user.Role, IsGuest: pgtype.Bool{Bool: user.Guest, Valid: true}})
		if err != nil {
			fmt.Println(err)
			return err
		}
	}

	err = dbQueries.InsertUserEvent(ctx, db.InsertUserEventParams{
		InternalMeetingID: meeting.InternalMeetingID,
		InternalUserID:    user.InternalUserID,
		EventType:         EventUserJoined,
		EventTimestamp:    pgtype.Timestamp{Time: b.GetTimestamp(), Valid: true},
	})
	if err != nil {
		fmt.Println(err)
		return err
	}
	return nil
}
