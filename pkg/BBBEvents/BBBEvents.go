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

import (
	db "bbbstatus/internal/database"
	"encoding/json"
	"fmt"
	"strings"
	"time"
)

// BaseEvent represents the common structure for all events
type BaseEvent struct {
	Data struct {
		Type       string     `json:"type"`
		ID         string     `json:"id"`
		Attributes Attributes `json:"attributes"`
		Event      struct {
			Timestamp int64 `json:"ts"`
		} `json:"event"`
	} `json:"data"`
}

// Attributes contains the common fields for all event attributes
type Attributes struct {
	Meeting Meeting      `json:"meeting"`
	User    *User        `json:"user,omitempty"`
	Poll    *Poll        `json:"poll,omitempty"`
	Message *ChatMessage `json:"chat-message,omitempty"`
	ChatID  string       `json:"chat-id,omitempty"`
}

// Meeting represents the meeting information present in all events
type Meeting struct {
	InternalMeetingID string  `json:"internal-meeting-id"`
	ExternalMeetingID string  `json:"external-meeting-id"`
	Name              string  `json:"name,omitempty"`
	IsBreakout        bool    `json:"is-breakout,omitempty"`
	ParentID          *string `json:"parent-id,omitempty"`
	CreateTimeStamp   int64   `json:"create-time,omitempty"`
	CreateTime        time.Time
	CreateDate        string            `json:"create-date,omitempty"`
	ModeratorPass     string            `json:"moderator-pass,omitempty"`
	ViewerPass        string            `json:"viewer-pass,omitempty"`
	Record            bool              `json:"record,omitempty"`
	VoiceConf         string            `json:"voice-conf,omitempty"`
	DialNumber        string            `json:"dial-number,omitempty"`
	MaxUsers          int               `json:"max-users,omitempty"`
	Metadata          map[string]string `json:"metadata,omitempty"`
	BbbHostname       string
	MeetingEnded      *time.Time
	ParticipantCount  int
}

func ConvertDBToBBBMeeting(dbMeeting db.Meeting) Meeting {
	// Handle ParentID conversion
	var parentID *string
	if dbMeeting.ParentID.Valid {
		parentID = &dbMeeting.ParentID.String
	}

	// Handle CreateTime conversion
	var createTime time.Time
	var createTimestamp int64
	if dbMeeting.CreateTime.Valid {
		createTime = dbMeeting.CreateTime.Time
		createTimestamp = createTime.UnixMilli()
	}

	// Handle MeetingEnded conversion
	var meetingEnded *time.Time
	if dbMeeting.MeetingEnded.Valid {
		meetingEnded = &dbMeeting.MeetingEnded.Time
	}

	// Convert Metadata from []byte to map[string]string
	metadata := make(map[string]string)
	if len(dbMeeting.Metadata) > 0 {
		json.Unmarshal(dbMeeting.Metadata, &metadata)
	}

	// Handle MaxUsers conversion
	maxUsers := 0
	if dbMeeting.MaxUsers.Valid {
		maxUsers = int(dbMeeting.MaxUsers.Int32)
	}

	// Handle ParticipantCount conversion
	participantCount := 0
	if dbMeeting.ParticipantCount.Valid {
		participantCount = int(dbMeeting.ParticipantCount.Int32)
	}

	// Handle VoiceConf conversion
	voiceConf := ""
	if dbMeeting.VoiceConf.Valid {
		voiceConf = dbMeeting.VoiceConf.String
	}

	// Handle DialNumber conversion
	dialNumber := ""
	if dbMeeting.DialNumber.Valid {
		dialNumber = dbMeeting.DialNumber.String
	}

	return Meeting{
		InternalMeetingID: dbMeeting.InternalMeetingID,
		ExternalMeetingID: dbMeeting.ExternalMeetingID,
		Name:              dbMeeting.Name,
		IsBreakout:        dbMeeting.IsBreakout.Bool,
		ParentID:          parentID,
		CreateTimeStamp:   createTimestamp,
		CreateTime:        createTime,
		CreateDate:        createTime.Format("2006-01-02"),
		ModeratorPass:     dbMeeting.ModeratorPass,
		ViewerPass:        dbMeeting.ViewerPass,
		Record:            dbMeeting.Record.Bool,
		VoiceConf:         voiceConf,
		DialNumber:        dialNumber,
		MaxUsers:          maxUsers,
		Metadata:          metadata,
		BbbHostname:       dbMeeting.Bbbhostname,
		MeetingEnded:      meetingEnded,
		ParticipantCount:  participantCount,
	}
}

// User represents a meeting participant
type User struct {
	InternalUserID string `json:"internal-user-id"`
	ExternalUserID string `json:"external-user-id"`
	Name           string `json:"name,omitempty"`
	Role           string `json:"role,omitempty"`
	Presenter      bool   `json:"presenter,omitempty"`
	Guest          bool   `json:"guest"`
	ListeningOnly  bool   `json:"listening-only,omitempty"`
	SharingMic     bool   `json:"sharing-mic,omitempty"`
	Muted          bool   `json:"muted,omitempty"`
	Stream         string `json:"stream,omitempty"`
	RaiseHand      bool   `json:"raise-hand,omitempty"`
	Emoji          string `json:"emoji,omitempty"`
	LeaveTimestamp *time.Time
}

// Poll represents a meeting poll
type Poll struct {
	ID        string   `json:"id"`
	Question  string   `json:"question,omitempty"`
	Answers   []Answer `json:"answers,omitempty"`
	AnswerIds []int    `json:"answerIds,omitempty"`
}

func (p Poll) ToString() string {
	var answers []string
	for _, answer := range p.Answers {
		answers = append(answers, answer.ToString())
	}
	return fmt.Sprintf("Poll(ID: %s, Question: %q, Answers: [%s], AnswerIds: %v)",
		p.ID, p.Question, strings.Join(answers, ", "), p.AnswerIds)
}

// Answer represents a poll answer option
type Answer struct {
	ID  int    `json:"id"`
	Key string `json:"key"`
}

func (a Answer) ToString() string {
	return fmt.Sprintf("Answer(ID: %d, Key: %q)", a.ID, a.Key)
}

// ChatMessage represents a chat message in the meeting
type ChatMessage struct {
	Message string `json:"message"`
	Sender  struct {
		InternalUserID string `json:"internal-user-id"`
		Name           string `json:"name"`
		Time           int64  `json:"time"`
	} `json:"sender"`
}

// GetTimestamp Helper function to convert timestamp to Time
func (b *BaseEvent) GetTimestamp() time.Time {
	return time.Unix(b.Data.Event.Timestamp/1000, 0)
}

type BBBEventType string

// Event type constants
const (
	EventMeetingCreated            = BBBEventType("meeting-created")
	EventUserJoined                = BBBEventType("user-joined")
	EventUserLeft                  = BBBEventType("user-left")
	EventUserPresenterAssigned     = BBBEventType("user-presenter-assigned")
	EventUserPresenterUnassigned   = BBBEventType("user-presenter-unassigned")
	EventUserAudioVoiceEnabled     = BBBEventType("user-audio-voice-enabled")
	EventUserAudioVoiceDisabled    = BBBEventType("user-audio-voice-disabled")
	EventUserAudioMuted            = BBBEventType("user-audio-muted")
	EventUserAudioUnmuted          = BBBEventType("user-audio-unmuted")
	EventUserCamBroadcastStart     = BBBEventType("user-cam-broadcast-start")
	EventUserCamBroadcastEnd       = BBBEventType("user-cam-broadcast-end")
	EventMeetingScreenshareStarted = BBBEventType("meeting-screenshare-started")
	EventMeetingScreenshareStopped = BBBEventType("meeting-screenshare-stopped")
	EventChatGroupMessageSent      = BBBEventType("chat-group-message-sent")
	EventUserEmojiChanged          = BBBEventType("user-emoji-changed")
	EventUserRaiseHandChanged      = BBBEventType("user-raise-hand-changed")
	EventPollStarted               = BBBEventType("poll-started")
	EventPollResponded             = BBBEventType("poll-responded")
	EventMeetingEnded              = BBBEventType("meeting-ended")
	EventMeetingRapArchiveEnded    = BBBEventType("meeting-rap-archive-ended") // TODO: use this as a counter for recordings of a meeting.
	EventMeetingRapArchiveStarted  = BBBEventType("meeting-rap-archive-started")
	EventSharedNotesChanged        = BBBEventType("pad-content")
	MeetingRecordingStarted        = BBBEventType("meeting-recording-started")
	MeetingRecordingStopped        = BBBEventType("meeting-recording-stopped")
)

var handledBBBEvents = []BBBEventType{EventMeetingCreated,
	EventUserJoined,
	EventUserLeft,
	EventUserPresenterAssigned,
	EventUserPresenterUnassigned,
	EventUserAudioVoiceEnabled,
	EventUserAudioVoiceDisabled,
	EventUserAudioMuted,
	EventUserAudioUnmuted,
	EventUserCamBroadcastStart,
	EventUserCamBroadcastEnd,
	EventMeetingScreenshareStarted,
	EventMeetingScreenshareStopped,
	EventChatGroupMessageSent,
	EventUserEmojiChanged,
	EventUserRaiseHandChanged,
	EventPollStarted,
	EventPollResponded,
	EventMeetingEnded,
	MeetingRecordingStarted,
	MeetingRecordingStopped,
}

/*
 TODO: Add rap-archive-started, rap-archive-ended, pad-content
samples:
{"data":{"type":"event","id":"rap-archive-started","attributes":{"meeting":{"internal-meeting-id":"06428c9e1f5f035ad02a0459fbf9a7e34b5c7384-1736508723885","external-meeting-id":"i9t0c1thxlupako4jxgygml8vpmipa4jsnl0lr1d"},"record-id":"06428c9e1f5f035ad02a0459fbf9a7e34b5c7384-1736508723885"},"event":{"ts":1736509334}}}
{"data":{"type":"event","id":"rap-archive-ended","attributes":{"meeting":{"internal-meeting-id":"06428c9e1f5f035ad02a0459fbf9a7e34b5c7384-1736508723885","external-meeting-id":"i9t0c1thxlupako4jxgygml8vpmipa4jsnl0lr1d"},"record-id":"06428c9e1f5f035ad02a0459fbf9a7e34b5c7384-1736508723885","success":true,"step-time":6841},"event":{"ts":1736509341}}}

*/
