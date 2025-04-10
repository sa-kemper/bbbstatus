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

func handleChatMessage(ctx context.Context, dbQueries *db.Queries, b *BaseEvent, meeting Meeting) (err error) {
	message := b.Data.Attributes.Message
	var userInRequestExists bool
	var senderID = b.Data.Attributes.Message.Sender.InternalUserID

	if senderID == "SYSTEM" {
		userInRequestExists = true

	} else {
		userInRequestExists, err = dbQueries.GetUserExistsByID(ctx, senderID)
		if err != nil {
			fmt.Println(err)
			return err
		}
	}

	if !userInRequestExists {
		return fmt.Errorf("user %s is not present. This is unexpected behaviour, check your DB integrity and the API authorisation", senderID)
	}
	err = dbQueries.InsertChatMessageToMeetingByID(ctx, db.InsertChatMessageToMeetingByIDParams{
		InternalMeetingID: meeting.InternalMeetingID,
		InternalUserID:    senderID,
		ChatID:            b.Data.Attributes.ChatID,
		MessageContent:    message.Message,
		SendTime:          pgtype.Timestamp{Valid: true, Time: b.GetTimestamp()},
	})
	if err != nil {
		fmt.Println(err)
		return err
	}
	return nil
}
