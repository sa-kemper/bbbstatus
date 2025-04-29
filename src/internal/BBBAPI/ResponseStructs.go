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
	RecordID          string            `xml:"recordID"`
	MeetingID         string            `xml:"meetingID"`
	InternalMeetingID string            `xml:"internalMeetingID"`
	Name              string            `xml:"name"`
	IsBreakout        bool              `xml:"isBreakout"`
	Published         bool              `xml:"published"`
	State             string            `xml:"state"`
	StartTime         int64             `xml:"startTime"`
	EndTime           int64             `xml:"endTime"`
	StartDate         time.Time         `xml:"startDate"`
	EndDate           time.Time         `xml:"endDate"`
	Participants      int               `xml:"participants"`
	MetaData          RecordingMetadata `xml:"metadata"`
	Playback          struct {
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

type GetMeetingsResponse struct {
	ReturnCode string      `xml:"returncode"`
	Meetings   AllMeetings `xml:"meetings"`
	MessageKey string      `xml:"messageKey"`
}
type AllMeetings struct {
	MeetingInfo []GetMeetingInfoResponse `xml:"meeting"`
}

type GetMeetingInfoResponse struct {
	ReturnCode            string    `xml:"returncode"`
	MeetingName           string    `xml:"meetingName"`
	MeetingID             string    `xml:"meetingID"`
	InternalMeetingID     string    `xml:"internalMeetingID"`
	CreateTime            string    `xml:"createTime"`
	CreateDate            string    `xml:"createDate"`
	VoiceBridge           string    `xml:"voiceBridge"`
	DialNumber            string    `xml:"dialNumber"`
	AttendeePW            string    `xml:"attendeePW"`
	ModeratorPW           string    `xml:"moderatorPW"`
	Running               bool      `xml:"running"`
	Duration              int       `xml:"duration"`
	HasUserJoined         bool      `xml:"hasUserJoined"`
	Recording             bool      `xml:"recording"`
	HasBeenForciblyEnded  bool      `xml:"hasBeenForciblyEnded"`
	StartTime             string    `xml:"startTime"`
	EndTime               string    `xml:"endTime"`
	ParticipantCount      int       `xml:"participantCount"`
	ListenerCount         int       `xml:"listenerCount"`
	VoiceParticipantCount int       `xml:"voiceParticipantCount"`
	VideoCount            int       `xml:"videoCount"`
	MaxUsers              int       `xml:"maxUsers"`
	ModeratorCount        int       `xml:"moderatorCount"`
	Attendees             attendees `xml:"attendees"`
	Metadata              string    `xml:"metadata"`
	MessageKey            string    `xml:"messageKey"`
	Message               string    `xml:"message"`
	//untested
	BreakoutRooms breakoutRooms `xml:"breakoutRooms"`
}
type breakoutRooms struct {
	BreakoutRooms []string `xml:"breakout"`
}

type attendees struct {
	Attendees []attendee `xml:"attendee"`
}

type attendee struct {
	UserID          string `xml:"userID"`
	FullName        string `xml:"fullName"`
	Role            string `xml:"role"`
	IsPresenter     bool   `xml:"isPresenter"`
	IsListeningOnly bool   `xml:"isListeningOnly"`
	HasJoinedVoice  bool   `xml:"hasJoinedVoice"`
	HasVideo        bool   `xml:"hasVideo"`
	Customdata      string `xml:"customdata"`
}
