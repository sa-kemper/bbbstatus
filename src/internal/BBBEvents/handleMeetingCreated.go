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
	"encoding/json"
	"fmt"
	"github.com/jackc/pgx/v5/pgtype"
)

func handleMeetingCreated(ctx context.Context, dbQueries *db.Queries, meeting Meeting, b *BaseEvent) error {
	var err error
	params := db.InsertMeetingParams{InternalMeetingID: meeting.InternalMeetingID,
		ExternalMeetingID: meeting.ExternalMeetingID,
		Name:              meeting.Name,
		IsBreakout:        pgtype.Bool{Bool: meeting.IsBreakout, Valid: true},
		CreateTime:        pgtype.Timestamp{Time: b.GetTimestamp(), Valid: true},
		ModeratorPass:     meeting.ModeratorPass,
		ViewerPass:        meeting.ViewerPass,
		Record:            pgtype.Bool{Bool: meeting.Record, Valid: true},
		VoiceConf:         pgtype.Text{String: meeting.VoiceConf, Valid: true},
		DialNumber:        pgtype.Text{String: meeting.DialNumber, Valid: true},
		MaxUsers:          pgtype.Int4{Int32: int32(meeting.MaxUsers), Valid: true},
		Bbbhostname:       meeting.BbbHostname,
	}

	if meeting.ParentID != nil {
		params.ParentID = pgtype.Text{String: *meeting.ParentID, Valid: true}
	} else {
		params.ParentID = pgtype.Text{String: "", Valid: false}
	}

	if meeting.VoiceConf != "" {
		params.VoiceConf = pgtype.Text{String: meeting.VoiceConf, Valid: true}
	} else {
		params.VoiceConf = pgtype.Text{String: "", Valid: false}
	}

	metadataBytes, err := json.Marshal(meeting.Metadata)
	if err != nil {
		fmt.Println("error marshalling meeting metadata")
		return err
	}
	params.Metadata = metadataBytes

	err = dbQueries.InsertMeeting(ctx, params)
	if err != nil {
		fmt.Println(err)
		return err
	}
	err = dbQueries.InsertMeetingEventForID(ctx, db.InsertMeetingEventForIDParams{InternalMeetingID: meeting.InternalMeetingID, EventType: EventMeetingCreated, EventTimestamp: pgtype.Timestamp{Time: b.GetTimestamp(), Valid: true}})
	if err != nil {
		fmt.Println(err)
		return err
	}
	return nil
}
