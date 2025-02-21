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

func handleMeetingCreated(tx pgx.Tx, meeting Meeting, b *BaseEvent) error {
	var err error
	_, err = tx.Exec(context.Background(),
		"INSERT INTO meetings (internal_meeting_id, external_meeting_id, name, is_breakout, parent_id,  create_time, moderator_pass, viewer_pass, record, voice_conf, dial_number, max_users, metadata, bbbhostname) VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13, $14)",
		meeting.InternalMeetingID,
		meeting.ExternalMeetingID,
		meeting.Name,
		meeting.IsBreakout,
		meeting.ParentID,
		b.GetTimestamp(),
		meeting.ModeratorPass,
		meeting.ViewerPass,
		meeting.Record,
		meeting.VoiceConf,
		meeting.DialNumber,
		meeting.MaxUsers,
		meeting.Metadata,
		meeting.BbbHostname,
	)
	if err != nil {
		fmt.Println(err)
		return err
	}
	_, err = tx.Exec(context.Background(), "INSERT INTO meeting_events (internal_meeting_id, event_type, event_timestamp) VALUES ($1, $2, $3)", meeting.InternalMeetingID, EventMeetingCreated, b.GetTimestamp())
	if err != nil {
		fmt.Println(err)
		return err
	}
	err = tx.Commit(context.Background())
	if err != nil {
		println(err)
		return err
	}
	return nil
}
