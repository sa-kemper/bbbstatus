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
	"net/http"
	"strconv"
)

type getRecordingsParameters struct {
	MeetingID *string
	RecordID  *string
	State     *string
	Meta      *string
	Offset    *int
	Limit     *int
}

func (a *API) GetRecordings(ctx context.Context, params getRecordingsParameters) (result Recordings, err error) {
	var client = a.getHTTPClient()
	var requestParameters = make(map[string]string)
	populateRequestParameters(requestParameters, params)

	request, err := http.NewRequestWithContext(ctx, "GET", generateURL(URLConfig{
		Hostname:     a.Hostname,
		Methode:      "getRecordings",
		Parameters:   requestParameters,
		SharedSecret: a.SharedSecret,
	}), nil)
	response, err := client.Do(request)
	if err != nil {
		return Recordings{}, err
	}

	err = xml.NewDecoder(response.Body).Decode(&result)
	if err != nil {
		return Recordings{}, err
	}

	return result, nil
}

func populateRequestParameters(paramMap map[string]string, params getRecordingsParameters) {
	if params.MeetingID != nil {
		paramMap["meeting_id"] = *params.MeetingID
	}
	if params.RecordID != nil {
		paramMap["record_id"] = *params.RecordID
	}
	if params.State != nil {
		paramMap["state"] = *params.State
	}
	if params.Meta != nil {
		paramMap["meta"] = *params.Meta
	}
	if params.Offset != nil {
		paramMap["offset"] = strconv.Itoa(*params.Offset)
	}
	if params.Limit != nil {
		paramMap["limit"] = strconv.Itoa(*params.Limit)
	}
}
