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
	"fmt"
	"net/http"
)

func (a *API) GetMeetings(ctx context.Context) (meetings []GetMeetingInfoResponse, err error) {
	var client = a.getHTTPClient()
	var requestParameters = make(map[string]string)
	var apiResponse GetMeetingsResponse
	// fmt.Println("DEBUG: GetMeetings -> a.Hostname = ", a.Hostname)
	var url = generateURL(URLConfig{Hostname: a.Hostname, Methode: "getMeetings", Parameters: requestParameters, SharedSecret: a.SharedSecret}) // we use the getMeetings methode as it should always respond with SUCCESS, even if no meetings exist.

	request, err := http.NewRequestWithContext(ctx, "GET", url, nil)
	if err != nil {
		fmt.Println("error occurred whilst building request (BBB API GetMeetings):", err)
		return meetings, err
	}

	response, err := client.Do(request)
	if err != nil {
		fmt.Println("error occurred whilst doing request (BBB API GetMeetings):", err)
		return meetings, err
	}

	err = xml.NewDecoder(response.Body).Decode(&apiResponse)
	if err != nil {
		fmt.Println("error occurred whilst decoding response (BBB API GetMeetings):", err)
		return meetings, err
	}

	if apiResponse.ReturnCode != "SUCCESS" {
		fmt.Println("error occurred whilst getting meetings:", apiResponse.MessageKey)
	}
	return apiResponse.Meetings.MeetingInfo, nil

}
