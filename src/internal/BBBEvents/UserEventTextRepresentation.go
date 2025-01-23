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

var UserEventTextRepresentation = map[string]i18n.Message{
	EventUserJoined:                {ID: "user-joined", Other: "The user '{{.Username}}' joined the meeting"},
	EventUserLeft:                  {ID: "user-left", Other: "The user '{{.Username}}' left the meeting"},
	EventUserPresenterAssigned:     {ID: "user-presenter-assigned", Other: "The user '{{.Username}}' has been assigned as the meetings presenter"},
	EventUserPresenterUnassigned:   {ID: "user-presenter-unassigned", Other: "The user '{{.Username}}' has been unassigned as the meetings presenter"},
	EventUserAudioVoiceEnabled:     {ID: "user-audio-voice-enabled", Other: "The user '{{.Username}}' has enabled joined the audio voice channel and can now hear others"},
	EventUserAudioVoiceDisabled:    {ID: "user-audio-voice-disabled", Other: "The user '{{.Username}} has left the audio voice channel and can no longer hear others"},
	EventUserAudioMuted:            {ID: "user-audio-muted", Other: "The user '{{.Username}}' muted himself and cannot be heard by others"},
	EventUserAudioUnmuted:          {ID: "user-audio-unmuted", Other: "The user '{{.Username}}' muted himself and cannot be heard by others"},
	EventUserCamBroadcastStart:     {ID: "user-cam-broadcast-start", Other: "The user '{{.Username}}' started a webcam broadcast"},
	EventUserCamBroadcastEnd:       {ID: "user-cam-broadcast-end", Other: "The user '{{.Username}}' ended a webcam broadcast"},
	EventMeetingScreenshareStarted: {ID: "meeting-screenshare-started", Other: "The user '{{.Username}}' started a screenshare"},
	EventMeetingScreenshareStopped: {ID: "meeting-screenshare-stopped", Other: "The user '{{.Username}}' stopped a screenshare"},
	EventUserEmojiChanged:          {ID: "chat-group-message-sent", Other: "The user '{{.Username}}' has changed his emoji"},
	EventUserRaiseHandChanged:      {ID: "user-emoji-changed", Other: "The user '{{.Username}}' raised his hand"},
}
