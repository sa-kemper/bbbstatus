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

func handleChatMessage(ctx context.Context, conn *pgx.Conn, b *BaseEvent, tx pgx.Tx, meeting Meeting) (err error) {
	message := b.Data.Attributes.Message
	var userInRequestExists bool
	userInRequestExists, err = userExists(ctx, conn, b.Data.Attributes.Message.Sender.InternalUserID)
	if err != nil {
		fmt.Println(err)
		return err
	}

	if !userInRequestExists {
		return fmt.Errorf("user %s is not present. This is unexpected behaviour, check your DB integrity and the API authorisation", b.Data.Attributes.Message.Sender.InternalUserID)
	}
	_, err = tx.Exec(context.Background(), "INSERT INTO chat_messages (internal_meeting_id, internal_user_id, chat_id, message_content, send_time) VALUES ($1,$2,$3,$4,$5)", meeting.InternalMeetingID, message.Sender.InternalUserID, b.Data.Attributes.ChatID, message.Message, b.GetTimestamp())
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
