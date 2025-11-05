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
	"net/http"
)

// ValidateApiSettings sends a simple request to the BBB server and throws an error if something goes wrong.
func (a *API) ValidateApiSettings(ctx context.Context) (valid bool, err error) {
	var client = a.getHTTPClient()
	var requestParameters = make(map[string]string)
	var apiResponse GetMeetingsResponse
	var url = generateURL(URLConfig{Hostname: a.Hostname, Methode: "getMeetings", Parameters: requestParameters, SharedSecret: a.SharedSecret}) // we use the getMeetings methode as it should always respond with SUCCESS, even if no meetings exist.
	request, err := http.NewRequestWithContext(ctx, "GET", url, nil)
	if err != nil {
		return false, err
	}

	response, err := client.Do(request)
	if err != nil {
		return false, err
	}

	err = xml.NewDecoder(response.Body).Decode(&apiResponse)
	if err != nil {
		return false, err
	}

	if apiResponse.ReturnCode == "SUCCESS" {
		return true, nil
	}
	return false, errors.New(apiResponse.ReturnCode + ":\t" + apiResponse.MessageKey)
}
