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

package BBBAPI

import (
	"context"
	"fmt"
)

func (a *API) GetServerUserCount(ctx context.Context) (userCount int, err error) {
	var meetings []GetMeetingInfoResponse
	meetings, err = a.GetMeetings(ctx)
	if err != nil {
		fmt.Println("error occurred while counting users for server (", a.Hostname, "): ", err)
		return
	}
	for _, meeting := range meetings {
		userCount += len(meeting.Attendees.Attendees)
	}
	return
}
