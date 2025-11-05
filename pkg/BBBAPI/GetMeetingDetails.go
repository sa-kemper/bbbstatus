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
	"encoding/xml"
	"errors"
	"fmt"
	"net/http"
)

// GetMeetingDetails gets information from the bbb server about the meeting in question
func (a *API) GetMeetingDetails(ctx context.Context, externalMeetingID string) (MeetingInfo GetMeetingInfoResponse, err error) {
	var client = a.getHTTPClient()
	var requestParameters = map[string]string{"meetingID": externalMeetingID}
	var url = generateURL(URLConfig{Hostname: a.Hostname, Methode: "getMeetingInfo", Parameters: requestParameters, SharedSecret: a.SharedSecret})
	request, err := http.NewRequestWithContext(ctx, "GET", url, nil)
	if err != nil {
		fmt.Println("error occurred when creating the API request (GetMeetingDetails):", err)
		return MeetingInfo, err
	}
	response, err := client.Do(request)
	if err != nil {
		fmt.Println("error occurred when doing the API request (GetMeetingDetails):", err)
		return MeetingInfo, err
	}
	defer response.Body.Close()

	err = xml.NewDecoder(response.Body).Decode(&MeetingInfo)
	if err != nil {
		fmt.Println("error occurred when decoding the API response (GetMeetingDetails):", err)
		return MeetingInfo, err
	}

	if MeetingInfo.ReturnCode != "SUCCESS" {
		fmt.Println("error occurred when doing the API request (GetMeetingDetails):", MeetingInfo.ReturnCode, MeetingInfo.Message)
		return MeetingInfo, errors.New(MeetingInfo.Message)
	}

	return MeetingInfo, nil
}
