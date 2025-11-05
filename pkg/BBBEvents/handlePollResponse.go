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
	"encoding/json"
	"fmt"

	"github.com/jackc/pgx/v5"
)

func handlePollResponse(b *BaseEvent, conn *pgx.Conn, user *User) error {
	poll := b.Data.Attributes.Poll
	if user == nil {
		return fmt.Errorf("user is unexpectedly nil")
	}
	res, err := json.Marshal(b.Data.Attributes.Poll.AnswerIds) // Convert struct to json, so it can be saved in the DB
	if err != nil {
		fmt.Println(err)
		return err
	}

	_, err = conn.Exec(context.Background(), "INSERT INTO poll_responses (poll_id, internal_user_id, answer_ids, response_time) VALUES ($1, $2, $3, $4)", poll.ID, user.InternalUserID, string(res), b.GetTimestamp())
	if err != nil {
		fmt.Println(err)
		return err
	}

	return nil
}
