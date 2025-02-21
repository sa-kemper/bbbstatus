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

func handleMeetingEnded(err error, tx pgx.Tx, meeting Meeting, b *BaseEvent) error {
	_, err = tx.Exec(context.Background(), "INSERT INTO meeting_events (internal_meeting_id, event_type, event_timestamp) VALUES ($1, $2, $3)", meeting.InternalMeetingID, EventMeetingEnded, b.GetTimestamp())
	if err != nil {
		fmt.Println(err)
		return err
	}
	_, err = tx.Exec(context.Background(), "UPDATE meetings SET meeting_ended = $1 WHERE internal_meeting_id = $2", b.GetTimestamp(), meeting.InternalMeetingID)
	if err != nil {
		fmt.Println("Failed ending the meeting.")
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
