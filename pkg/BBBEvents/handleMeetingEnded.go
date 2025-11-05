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

func handleMeetingEnded(ctx context.Context, dbQueries *db.Queries, meeting Meeting, b *BaseEvent) (err error) {
	err = dbQueries.InsertMeetingEventForID(ctx, db.InsertMeetingEventForIDParams{
		InternalMeetingID: meeting.InternalMeetingID,
		EventType:         EventMeetingEnded,
		EventTimestamp:    pgtype.Timestamp{Valid: true, Time: b.GetTimestamp()},
	})

	if err != nil {
		fmt.Println(err)
		return err
	}
	err = dbQueries.EndMeetingAtTimestampByID(ctx, db.EndMeetingAtTimestampByIDParams{
		MeetingEnded:      pgtype.Timestamp{Valid: true, Time: b.GetTimestamp()},
		InternalMeetingID: meeting.InternalMeetingID,
	})
	if err != nil {
		fmt.Println("Failed ending the meeting.")
		fmt.Println(err)
		return err
	}
	return nil
}
