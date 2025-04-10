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

func handlePollCreation(ctx context.Context, dbQueries *pgx.Conn, b *BaseEvent, user *User, meeting Meeting) (err error) {
	poll := b.Data.Attributes.Poll
	if user == nil {
		return fmt.Errorf("user is unexpectedly nil")
	}
	_, err = dbQueries.Exec(ctx, "INSERT INTO polls (poll_id, internal_meeting_id,  internal_user_id, question, answers, created_at) VALUES ($1, $2, $3, $4, $5, $6)", poll.ID, meeting.InternalMeetingID, user.InternalUserID, poll.Question, poll.Answers, b.GetTimestamp())
	if err != nil {
		fmt.Println(err)
		return err
	}
	return nil
}
