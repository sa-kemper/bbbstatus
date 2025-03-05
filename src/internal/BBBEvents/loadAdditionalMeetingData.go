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

func loadAdditionalMeetingData(meeting Meeting, tx pgx.Tx) error {
	if meeting.Name == "" && meeting.InternalMeetingID != "" {
		err := tx.QueryRow(context.Background(), "SELECT name FROM meetings WHERE internal_meeting_id=$1", meeting.InternalMeetingID).Scan(&meeting.Name)
		if err != nil {
			fmt.Println("error occurred during save event -> loadAdditionalMeetingData: ", err)
			return err
		}
	}
	return nil
}
