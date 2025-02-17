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

import "time"

type GetRecordingsResponse struct {
	ReturnCode string     `xml:"returncode"`
	Recordings Recordings `xml:"recordings"`
}
type Recordings struct {
	Recording []Recording `xml:"recording"`
}
type Recording struct {
	RecordID     string            `xml:"recordID"`
	MeetingID    string            `xml:"meetingID"`
	Name         string            `xml:"name"`
	IsBreakout   bool              `xml:"isBreakout"`
	Published    string            `xml:"published"`
	State        string            `xml:"state"`
	StartTime    time.Time         `xml:"startTime"`
	EndTime      time.Time         `xml:"endTime"`
	Participants int               `xml:"participants"`
	MetaData     RecordingMetadata `xml:"metadata"`
	Playback     struct {
		Format []struct {
			Type   string   `xml:"type"`
			Url    string   `xml:"url"`
			Length string   `xml:"length"`
			Images []string `xml:"preview>images>image"`
		} `xml:"format"`
	} `xml:"playback"`
}

type RecordingMetadata struct {
	IsBreakout       bool   `xml:"isBreakout"`
	MeetingName      string `xml:"meetingName"`
	GreenLightListed bool   `xml:"gl-listed"`
	MeetingId        string `xml:"meetingId"`
}
