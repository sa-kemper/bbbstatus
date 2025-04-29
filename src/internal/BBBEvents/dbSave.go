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
	"errors"
	"fmt"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgtype"
	"slices"
	"strings"
)

func (b *BaseEvent) Save(ctx context.Context, dbQueries *db.Queries, conn *pgx.Conn) error {
	var err error
	var meeting = b.Data.Attributes.Meeting
	var user *User

	if b.Data.Attributes.User != nil {
		user = b.Data.Attributes.User
	}

	if !slices.Contains(handledBBBEvents, b.Data.ID) {
		fmt.Println("WARNING:", b.Data.ID, "is not being handled by bbbstatus yet.")
		return nil
	}

	if isUserEvent(b.Data.ID) {
		if slices.Contains([]string{EventMeetingScreenshareStarted, EventMeetingScreenshareStopped}, b.Data.ID) {
			dbUser, err := dbQueries.GetPresenterUserByMeetingID(ctx, meeting.InternalMeetingID)
			if err != nil {
				fmt.Println("error occured whilst loadAdditionalUserData -> get meeting presenter: ", err)
				return err
			}
			user = &User{InternalUserID: dbUser.InternalUserID, ExternalUserID: dbUser.ExternalUserID, Name: dbUser.Name, Role: dbUser.Role, Presenter: true, Guest: dbUser.IsGuest.Bool}
		}

		err = loadAdditionalUserData(ctx, user, dbQueries)
		if err != nil {
			if errors.Is(err, pgx.ErrNoRows) {
				// bbb-webhooks messed up the queue order again

			}
		}

		if b.Data.ID == EventUserPresenterAssigned { // This is a special case because bbb-webhooks does not maintain the event queue order, if one event is not delivered. (Throw away presenter assigned events if the user is unknown)
			if errors.Is(err, pgx.ErrNoRows) {
				return nil
			}
		}

		if err != nil {
			fmt.Println("error occurred during save event -> loadAdditionalUserData: ", err)
			return err
		}
	}

	if b.Data.ID != EventMeetingCreated {
		err = loadAdditionalMeetingData(ctx, &meeting, dbQueries)
		if err != nil {
			fmt.Println("error occurred during save event -> loadAdditionalMeetingData: ", err)
			return err
		}
	}

	if isUserEvent(b.Data.ID) {
		err = handleUserEvent(ctx, dbQueries, user, meeting, b)
		if err != nil {
			fmt.Println("error occurred during save event -> handleUserEvent: ", err)
		}
	}

	switch b.Data.ID {
	case EventMeetingCreated:
		return handleMeetingCreated(ctx, dbQueries, meeting, b)
	case EventChatGroupMessageSent:
		return handleChatMessage(ctx, dbQueries, b, meeting)
	case EventPollStarted: //
		return handlePollCreation(ctx, conn, b, user, meeting)
	case EventPollResponded:
		return handlePollResponse(b, conn, user)
	case EventMeetingEnded:
		return handleMeetingEnded(ctx, dbQueries, meeting, b)
	case MeetingRecordingStarted:
		return handleMeetingRecordEvent(ctx, dbQueries, meeting, b)
	case MeetingRecordingStopped:
		return handleMeetingRecordEvent(ctx, dbQueries, meeting, b)
	case EventMeetingRapArchiveEnded:
		return handleRecordingArchivedEvent(ctx, dbQueries, meeting, b)
	}
	return nil

}

func handleRecordingArchivedEvent(ctx context.Context, queries *db.Queries, meeting Meeting, b *BaseEvent) (err error) {
	if b.Data.ID == "rap-archive-ended" { // when a recording is processed this event is sent.
		err = queries.IncrementRecordingsCountForServer(ctx, meeting.BbbHostname)
		if err != nil {
			fmt.Println("error occurred during IncrementRecordingsCountForServer: ", err, "Server: ", b.Data.Attributes.Meeting.BbbHostname)
			return err
		}
	}
	return nil
}

// isUserEvent determines if the given event ID should be handled as a user event.
// It includes events containing "user" and specific meeting-related events.
func isUserEvent(eventID string) bool {
	if strings.Contains(eventID, "user") {
		return true
	}
	switch eventID {
	case EventMeetingScreenshareStarted, EventMeetingScreenshareStopped:
		return true
	default:
		return false
	}
}

func handleMeetingRecordEvent(ctx context.Context, dbQueries *db.Queries, meeting Meeting, b *BaseEvent) (err error) {
	err = dbQueries.InsertMeetingEventForID(ctx, db.InsertMeetingEventForIDParams{InternalMeetingID: meeting.InternalMeetingID, EventType: b.Data.ID, EventTimestamp: pgtype.Timestamp{Valid: true, Time: b.GetTimestamp()}})
	if err != nil {
		fmt.Println("error occurred during save event -> handleMeetingRecordEvent: ", err)
		return err
	}
	return nil
}
