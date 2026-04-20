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

import "github.com/nicksnyder/go-i18n/v2/i18n"

var UserEventTextRepresentation = map[BBBEventType]i18n.Message{
	EventUserJoined:                {ID: "user-joined", Other: "User '{{.Username}}' joined the meeting"},
	EventUserLeft:                  {ID: "user-left", Other: "User '{{.Username}}' left the meeting"},
	EventUserPresenterAssigned:     {ID: "user-presenter-assigned", Other: "User '{{.Username}}' has been assigned as the meetings presenter"},
	EventUserPresenterUnassigned:   {ID: "user-presenter-unassigned", Other: "User '{{.Username}}' has been unassigned as the meetings presenter"},
	EventUserAudioVoiceEnabled:     {ID: "user-audio-voice-enabled", Other: "User '{{.Username}}' has enabled joined the audio voice channel and can now hear others"},
	EventUserAudioVoiceDisabled:    {ID: "user-audio-voice-disabled", Other: "User '{{.Username}} has left the audio voice channel and can no longer hear others"},
	EventUserAudioMuted:            {ID: "user-audio-muted", Other: "User '{{.Username}}' muted himself and cannot be heard by others"},
	EventUserAudioUnmuted:          {ID: "user-audio-unmuted", Other: "User '{{.Username}}' muted himself and cannot be heard by others"},
	EventUserCamBroadcastStart:     {ID: "user-cam-broadcast-start", Other: "The user '{{.Username}}' started a webcam broadcast"},
	EventUserCamBroadcastEnd:       {ID: "user-cam-broadcast-end", Other: "The user '{{.Username}}' ended a webcam broadcast"},
	EventMeetingScreenshareStarted: {ID: "meeting-screenshare-started", Other: "User '{{.Username}}' started a screenshare"},
	EventMeetingScreenshareStopped: {ID: "meeting-screenshare-stopped", Other: "User '{{.Username}}' stopped a screenshare"},
	EventUserEmojiChanged:          {ID: "user-emoji-changed", Other: "User '{{.Username}}' has changed his emoji"},
	EventUserRaiseHandChanged:      {ID: "user-raise-hand-changed", Other: "User '{{.Username}}' raised his hand"},
}
