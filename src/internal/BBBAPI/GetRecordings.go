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
	"io"
	"net/http"
	"strconv"
)

// GetRecordingsParameters holds the parameters for fetching recordings from the API.
type GetRecordingsParameters struct {
	// MeetingID is a meeting ID for getting the recordings. It can be a set of meetingIDs separated by commas. If the meeting ID is not specified, it will get ALL the recordings. If a recordID is specified, the meetingID is ignored.
	MeetingID *string

	// RecordID is a record ID for getting the recordings. It can be a set of recordIDs separated by commas. If the record ID is not specified, it will use meeting ID as the main criteria. If neither the meeting ID nor record ID is specified, it will get ALL the recordings. The recordID can also be used as a wildcard by including only the first characters in the string.
	RecordID *string

	// State indicates the state of the recording. It can be one of [processing|processed|published|unpublished|deleted]. The parameter state can be used to filter results. If it is not specified, only the states [published|unpublished] are considered. If specified as “any”, recordings in all states are included.
	State *string

	// Meta allows passing one or more metadata values to filter the recordings returned. The format of these parameters is the same as the metadata passed to the create call.
	Meta *string

	// Offset is the starting index for returned recordings. The number must be greater than or equal to 0.
	Offset *int

	// Limit is the maximum number of recordings to be returned. The number must be between 1 and 100.
	Limit *int
}

func (a *API) GetRecordings(ctx context.Context, params GetRecordingsParameters) (result Recordings, err error) {
	var client = a.getHTTPClient()
	var requestParameters = make(map[string]string)
	var apiResponse GetRecordingsResponse
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

	defer response.Body.Close()
	responseBytes, err := io.ReadAll(response.Body)
	if err != nil {
		return Recordings{}, err
	}

	err = xml.Unmarshal(responseBytes, &apiResponse)
	if err != nil {
		return Recordings{}, err
	}

	return apiResponse.Recordings, nil
}

func populateRequestParameters(paramMap map[string]string, params GetRecordingsParameters) {
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
