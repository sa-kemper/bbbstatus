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

import "testing"

func TestGenerateURL(t *testing.T) {
	type args struct {
		config URLConfig
	}
	tests := []struct {
		name string
		args args
		want string
	}{
		{
			name: "Generate Valid Create URL",
			args: args{
				config: URLConfig{
					Hostname: "test-install.blindsidenetworks.com",
					Methode:  "create",
					Parameters: map[string]string{
						"allowStartStopRecording": "true",
						"attendeePW":              "AttendeePassword",
						"autoStartRecording":      "false",
						"meetingID":               "TestMeetingID",
						"moderatorPW":             "ModeratorPassword",
						"name":                    "TestMeetingID",
						"record":                  "false",
						"voiceBridge":             "71225",
					},
					SharedSecret: "8cd8ef52e8e101574e400365b55e11a6",
				},
			},
			want: "https://test-install.blindsidenetworks.com/bigbluebutton/api/create?allowStartStopRecording=true&attendeePW=AttendeePassword&autoStartRecording=false&meetingID=TestMeetingID&moderatorPW=ModeratorPassword&name=TestMeetingID&record=false&voiceBridge=71225&checksum=cef439299383b3430967b5277ffec13c8e310001",
		},
		{
			name: "Generate Valid GetMeetings URL",
			args: args{
				config: URLConfig{
					Hostname:     "test-install.blindsidenetworks.com",
					Methode:      "getMeetings",
					Parameters:   map[string]string{},
					SharedSecret: "8cd8ef52e8e101574e400365b55e11a6",
				},
			},
			want: "https://test-install.blindsidenetworks.com/bigbluebutton/api/getMeetings?checksum=d23fef405937517be465ffccae12d5c1103a5e00",
		},
		{
			name: "Generate Valid getRecordings URL",
			args: args{
				config: URLConfig{
					Hostname: "test-install.blindsidenetworks.com",
					Methode:  "getRecordings",
					Parameters: map[string]string{
						"meetingID": "TestMeetingID",
					},
					SharedSecret: "8cd8ef52e8e101574e400365b55e11a6",
				},
			},
			want: "https://test-install.blindsidenetworks.com/bigbluebutton/api/getRecordings?meetingID=TestMeetingID&checksum=765af7b0168adb286f3e07ee154d11f562b6a373",
		},
		{
			name: "Generate Valid create URL",
			args: args{
				config: URLConfig{
					Hostname: "bbb.example.com",
					Methode:  "create",
					Parameters: map[string]string{
						"allowStartStopRecording": "true",
						"attendeePW":              "AttendeePassword",
						"autoStartRecording":      "false",
						"meetingID":               "TestMeetingID",
						"moderatorPW":             "ModeratorPassword",
						"name":                    "TestMeetingID",
						"record":                  "false",
						"voiceBridge":             "71225",
					},
					SharedSecret: "yourcoolsharedsecretandall123",
				},
			},
			want: "https://bbb.example.com/bigbluebutton/api/create?allowStartStopRecording=true&attendeePW=AttendeePassword&autoStartRecording=false&meetingID=TestMeetingID&moderatorPW=ModeratorPassword&name=TestMeetingID&record=false&voiceBridge=71225&checksum=04caf6b3a2366befad71b9a5cb3e0c2178654abd",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := generateURL(tt.args.config); got != tt.want {
				t.Errorf("generateURL() = %v, want %v", got, tt.want)
			}
		})
	}
}
