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

func handleUserJoined(ctx context.Context, conn *pgx.Conn, user *User, meeting Meeting, b *BaseEvent) error {
	var err error
	var userInRequestExists bool

	if user == nil {
		return fmt.Errorf("user is unexpectedly nil")
	}

	userInRequestExists, err = userExists(ctx, conn, user.InternalUserID)
	if err != nil {
		fmt.Println(err)
		return err
	}

	if !userInRequestExists {
		_, err := conn.Exec(context.Background(), "INSERT INTO users (internal_user_id, external_user_id, name, role, is_guest) VALUES ($1, $2, $3, $4, $5)", user.InternalUserID, user.ExternalUserID, user.Name, user.Role, user.Guest)
		if err != nil {
			fmt.Println(err)
			return err
		}
	}

	_, err = conn.Exec(context.Background(), "INSERT INTO user_events (internal_meeting_id, internal_user_id, event_type, event_timestamp) VALUES ($1, $2, $3, $4)", meeting.InternalMeetingID, user.InternalUserID, b.Data.ID, b.GetTimestamp())
	if err != nil {
		fmt.Println(err)
		return err
	}
	return nil
}
