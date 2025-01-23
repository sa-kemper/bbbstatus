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
	"strings"
)

func userExists(ctx context.Context, conn *pgx.Conn, InternalUserID string) (exists bool, err error) {
	err = conn.QueryRow(ctx, "SELECT COUNT(*) > 0 AS user_exists FROM users WHERE internal_user_id = $1", InternalUserID).Scan(&exists)
	if err != nil {
		fmt.Println(err)
		return false, err
	}
	return exists, nil
}

func (b *BaseEvent) Save(ctx context.Context, conn *pgx.Conn) error {
	var err error
	var meeting = b.Data.Attributes.Meeting
	var user *User

	if b.Data.Attributes.User != nil {
		user = b.Data.Attributes.User
	}

	tx, err := conn.Begin(ctx) // We use database transactions to handle runtime errors gracefully.
	if err != nil {
		fmt.Println(err)
		return err
	}
	//goland:noinspection ALL
	defer tx.Rollback(ctx)

	err2 := loadAdditionalUserData(user, tx)
	if err2 != nil {
		return err2
	}

	err3 := loadAdditionalMeetingData(meeting, tx)
	if err3 != nil {
		return err3
	}

	if isUserEvent := strings.Contains(b.Data.ID, "user"); isUserEvent {
		return handleUserEvent(ctx, conn, user, tx, meeting, b)
	}
	if b.Data.ID == EventMeetingScreenshareStarted || b.Data.ID == EventMeetingScreenshareStopped {
		return handleUserEvent(ctx, conn, user, tx, meeting, b)
	}

	switch b.Data.ID {
	case EventMeetingCreated:
		return handleMeetingCreated(tx, meeting, b)
	case EventChatGroupMessageSent:
		return handleChatMessage(ctx, conn, b, err, tx, meeting)
	case EventPollStarted: //
		return handlePollCreation(b, user, err, tx, meeting)
	case EventPollResponded:
		return handlePollResponse(b, user, tx)
	case EventMeetingEnded:
		return handleMeetingEnded(err, tx, meeting, b)
	}
	return nil

}
