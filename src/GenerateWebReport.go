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

package main

import (
	"bbbstatus/internal/BBBAPI"
	"bbbstatus/internal/BBBEvents"
	db "bbbstatus/internal/database"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"github.com/jackc/pgx/v5"
	"github.com/nicksnyder/go-i18n/v2/i18n"
	"os"
	"sort"
	"time"
)

type Detail struct {
	Name  string
	Value string
}

type Event struct {
	Time               time.Time
	FormattedTime      string
	TextRepresentation string
}

type Report struct {
	Details      []Detail
	Participants []BBBEvents.User
	Timeline     []Event
	Recordings   []BBBAPI.Recording
}

func init() { // Add all messages that are related to this file into the localization bundle
	var msgs = []i18n.Message{
		{ID: "EventTimelineHeader", Other: "Events Timeline"},
		{ID: "ReportMeetingDetailMeetingName", Other: "Meeting name"},
		{ID: "ReportHeader", Other: "Meeting Report"},
		{ID: "ReportMeetingDetailsHeader", Other: "Meeting Details"},
		{ID: "ReportParticipantsHeader", Other: "Participants"},
		{ID: "ReportPollResponseEventRepresentation", Other: "The user '{{.Username}}' responded to a poll {{.PollQuestion}} with {{.PollAnswer}}"},
		{ID: "ReportPollStartedEventRepresentation", Other: "The user '{{.Username}}' started a poll '{{.PollQuestion}}', with options: {{.PollOptions}}"},
		{ID: "ReportMessageEventRepresentation", Other: "The user '{{.Username}}' sent a message: {{.Message}}"},
		{ID: "ReportPrintReportButton", Other: "Print report"},
		{ID: "ReportOpenInExcelButton", Other: "Open in excel"},
		{ID: "ReportMeetingDetailInternalMeetingID", Other: "Internal Meeting ID"},
		{ID: "ReportMeetingDetailBBBHostname", Other: "BBB Hostname"},
		{ID: "ReportMeetingDetailCreationDate", Other: "Creation Date"},
		{ID: "meetingListHeader", Other: "Meeting List"},
		{ID: "BackToMeetingsButton", Other: "Back to Meetings"},
		{ID: "RecordingsHeader", Other: "Recordings"},
		{ID: "SystemSentMessage", Other: "System Sent the message: '{{.Message}}'"},
		{ID: "MeetingEvent-meeting-created", Other: "Meeting created"},
		{ID: "MeetingEvent-meeting-ended", Other: "Meeting ended"},
		{ID: "MeetingEvent-meeting-recording-started", Other: "Meeting Recording Started"},
		{ID: "MeetingEvent-meeting-recording-stopped", Other: "Meeting Recording Stopped"},
	}
	FrontendTextMessages = append(FrontendTextMessages, msgs...)
	for _, m := range BBBEvents.UserEventTextRepresentation { // Add user events text representation to the language strings.
		FrontendTextMessages = append(FrontendTextMessages, m)
	}
}

func GenerateWebReport(ctx context.Context, internalMeetingID string) (Report, error) {
	var meeting BBBEvents.Meeting
	var details []Detail
	var participants []BBBEvents.User
	var timeline []Event
	var polls []string
	var meetingServerAPI BBBAPI.API
	var recordings []BBBAPI.Recording

	// Connect to the db using
	conn, err := pgx.Connect(ctx, confGet("DB_CONNECTION_STRING"))
	if err != nil {
		return Report{}, fmt.Errorf("Unable to connect to database: %v\n", err)
	}
	//goland:noinspection ALL
	defer conn.Close(ctx)

	dbQueries := db.New(conn)

	// Query and parse meeting using the row.next and row.scan methode of pgx
	err = conn.QueryRow(ctx, "SELECT * FROM meetings WHERE internal_meeting_id = $1 LIMIT 1", internalMeetingID).Scan(
		&meeting.InternalMeetingID,
		&meeting.ExternalMeetingID,
		&meeting.Name,
		&meeting.IsBreakout,
		&meeting.ParentID,
		&meeting.CreateTime,
		&meeting.ModeratorPass,
		&meeting.ViewerPass,
		&meeting.Record,
		&meeting.VoiceConf,
		&meeting.DialNumber,
		&meeting.MaxUsers,
		&meeting.Metadata,
		&meeting.BbbHostname,
		&meeting.ParticipantCount,
		&meeting.MeetingEnded,
	)
	if err != nil {
		return Report{}, fmt.Errorf("Unable to find meeting with Id %s: %v\n", internalMeetingID, err)
	}

	servers := confGetServers(meeting.BbbHostname)
	var server bbbServer

	if len(servers) < 1 {
		panic("Unable to find BBBServer for meeting " + internalMeetingID)
	} else {
		server = servers[0]
		if server.Hostname == meeting.BbbHostname {
			if server.APITimeout != 0 {
				apitimeout := time.Duration(server.APITimeout) * time.Second
				meetingServerAPI = BBBAPI.API{Hostname: server.Hostname, Port: server.ApiPort, SharedSecret: server.SharedSecret, Timeout: &apitimeout}
			} else {
				meetingServerAPI = BBBAPI.API{Hostname: server.Hostname, Port: server.ApiPort, SharedSecret: server.SharedSecret}
			}
		} else {
			//fmt.Println("DEBUG API SERVER FINDING: '" + server.Hostname + "' != '" + meeting.BbbHostname + "'")
		}
	}
	//fmt.Println("DEBUG: meetingAPI: ", meetingServerAPI)
	details = FillMeetingDetails(details, meeting)

	err, participants = FillMeetingParticipants(ctx, internalMeetingID, conn)
	if err != nil {
		return Report{}, err
	}

	if meetingServerAPI.Hostname != "" && meetingServerAPI.SharedSecret != "" {
		//fmt.Println("DEBUG: meetingServerAPI is valid enough ^^ ", meetingServerAPI)
		getRecordingsResponse, err := meetingServerAPI.GetRecordings(ctx, BBBAPI.GetRecordingsParameters{MeetingID: &meeting.ExternalMeetingID}, meeting)
		if err != nil {
			if errors.Is(err, os.ErrDeadlineExceeded) {
				fmt.Println("FATAL: BBB API Timeout for host:", meetingServerAPI.Hostname)
			}
			//fmt.Println("DEBUG: Error occurred whilst geting recordings. ", err)
			return Report{}, err
		}
		recordings = getRecordingsResponse.Recording
		//fmt.Println("DEBUG: recordings: ", recordings)
	}

	userEvents, err := FillMeetingUserEvents(ctx, internalMeetingID, conn)
	if err != nil {
		return Report{}, err
	}
	timeline = append(timeline, userEvents...)

	messageTimeline, err := FillMeetingMessageEvents(ctx, internalMeetingID, conn)
	if err != nil {
		return Report{}, err
	}
	timeline = append(timeline, messageTimeline...)

	pollTimeline, err := FillMeetingPollEvents(ctx, internalMeetingID, conn, &polls)
	if err != nil {
		return Report{}, err
	}
	timeline = append(timeline, pollTimeline...)

	pollResponseTimeline, err := FillMeetingPollResponses(ctx, internalMeetingID, &polls, conn)
	if err != nil {
		return Report{}, err
	}
	timeline = append(timeline, pollResponseTimeline...)

	meetingEventsTimeline, err := FillMeetingEvents(ctx, internalMeetingID, dbQueries)
	if err != nil {
		return Report{}, err
	}
	timeline = append(timeline, meetingEventsTimeline...)

	for i, event := range timeline {
		timeline[i].FormattedTime = event.Time.Format(time.TimeOnly + " 02.01.2006")
	}

	// sort the timeline
	sort.Slice(timeline, func(i, j int) bool {
		return timeline[i].Time.Before(timeline[j].Time)
	})

	return Report{Details: details, Participants: participants, Recordings: recordings, Timeline: timeline}, nil
}

func FillMeetingEvents(ctx context.Context, id string, dbQueries *db.Queries) ([]Event, error) {
	meetingEvents, err := dbQueries.GetMeetingEventsByInternalMeetingID(ctx, id)
	if err != nil {
		return nil, fmt.Errorf("failed to get meeting events for meeting id %s: %w", id, err)
	}

	timeline := make([]Event, 0, len(meetingEvents))
	for _, event := range meetingEvents {
		if !event.EventTimestamp.Valid {
			fmt.Println("WARNING: Meeting event with id ", id, " has an invalid timestamp, event type:", event.EventType)
			continue // Skip invalid timestamps; log if needed
		}
		timeline = append(timeline, Event{
			Time:               event.EventTimestamp.Time,
			TextRepresentation: Translate("MeetingEvent-" + event.EventType),
		})
	}
	return timeline, nil
}

func FillMeetingPollResponses(ctx context.Context, internalMeetingID string, polls *[]string, conn *pgx.Conn) (timeline []Event, err error) {
	// insert the poll Answers into the timeline
	for _, pollID := range *polls {
		row, err := conn.Query(ctx, "SELECT internal_user_id, answer_ids, response_time FROM poll_responses WHERE poll_id = $1", pollID)
		if err != nil {
			return timeline, fmt.Errorf("Unable to find meeting with Id %s: %v\n", internalMeetingID, err)
		}
		for row.Next() {
			var event Event
			var poll = BBBEvents.Poll{ID: pollID}
			var user = BBBEvents.User{}
			var answerJSON string
			var answersJSON string

			var err = row.Scan(&user.InternalUserID, &answerJSON, &event.Time)
			if err != nil {
				fmt.Println(err)
			}

			err = json.Unmarshal([]byte(answerJSON), &poll.AnswerIds)
			if err != nil {
				fmt.Printf("error decoding answers of poll %s with json '%s'\n", poll.ID, answerJSON)
				return timeline, err
			}

			// we cannot use the same connection while the row is not closed.
			conn2, err := pgx.Connect(ctx, confGet("DB_CONNECTION_STRING"))
			if err != nil {
				fmt.Println("error connecting to the database at least twice. ->", err)
				return timeline, err
			}
			err = conn2.QueryRow(ctx, "SELECT name FROM users WHERE internal_user_id = $1", user.InternalUserID).Scan(&user.Name)
			if err != nil {
				fmt.Println("error obtaining user information of poll answer")
				fmt.Println(err)
				user.Name = user.InternalUserID
			}

			err = conn2.QueryRow(ctx, "SELECT answers FROM polls WHERE poll_id = $1", pollID).Scan(&answersJSON)
			if err != nil {
				fmt.Println("error obtaining answers ->", err)
			}

			err = json.Unmarshal([]byte(answersJSON), &poll.Answers)
			if err != nil {
				fmt.Printf("error decoding answers of poll %s with json '%s'\n", poll.ID, answerJSON)

			}
			if poll.Question == "" { // use a shortened version of the poll id if the question is empty.
				poll.Question = string([]byte(pollID)[0:5])
			}
			event.TextRepresentation = TranslateAdvanced("ReportPollResponseEventRepresentation", map[string]string{"username": user.Name, "pollQuestion": poll.Question, "pollAnswer": poll.Answers[poll.AnswerIds[0]].Key}) // fmt.Sprintf("%s responded to a poll '%s' with %s", user.Name, poll.Question, poll.Answers[poll.AnswerIds[0]].Key)
			timeline = append(timeline, event)
			//goland:noinspection ALL
			conn2.Close(ctx)

		}
		row.Close()
	}
	return timeline, err
}

func FillMeetingPollEvents(ctx context.Context, internalMeetingID string, conn *pgx.Conn, polls *[]string) (timeline []Event, err error) {
	// insert the polls into the timeline
	row, err := conn.Query(ctx, "SELECT poll_id, internal_user_id, question, answers, created_at FROM polls WHERE internal_meeting_id = $1", internalMeetingID)
	if err != nil {
		return timeline, fmt.Errorf("Unable to find meeting with Id %s: %v\n", internalMeetingID, err)
	}
	for row.Next() {
		var event Event
		var pollId string
		var userID string
		var userName string
		var question string
		var answers string
		var answersTextRepresentation string

		err = row.Scan(&pollId, &userID, &question, &answers, &event.Time)
		if err != nil {
			fmt.Println(err)
		}
		*polls = append(*polls, pollId)

		var answerObjects []BBBEvents.Answer
		err = json.Unmarshal([]byte(answers), &answerObjects)
		for _, answer := range answerObjects {
			answersTextRepresentation += fmt.Sprintf("'%s', ", answer.Key)
		}

		// we cannot use the same connection while the row is not closed.
		conn2, err := pgx.Connect(ctx, confGet("DB_CONNECTION_STRING"))
		if err != nil {
			fmt.Println("error connecting to the database at least twice. ->", err)
			return timeline, err
		}
		err = conn2.QueryRow(ctx, "SELECT name FROM users WHERE internal_user_id = $1", userID).Scan(&userName)
		if err != nil {
			fmt.Println("error obtaining user information of chat message")
			fmt.Println(err)
		}

		if question != "" { // make a suitable question replacement in case the user drops the question content altogether.
			question = string([]byte(pollId)[0:5])
		}

		event.TextRepresentation = TranslateAdvanced("ReportPollStartedEventRepresentation", map[string]string{"Username": userName, "PollQuestion": question, "PollOptions": answersTextRepresentation}) // fmt.Sprintf("%s started a poll '%s', with options: %s", userName, question, answersTextRepresentation)
		timeline = append(timeline, event)
		//goland:noinspection ALL
		conn2.Close(ctx)

	}
	row.Close()
	return timeline, err
}

func FillMeetingMessageEvents(ctx context.Context, internalMeetingID string, dbQueries *db.Queries) (timeline []Event, err error) {
	// insert user messages into the timeline
	meetingMessages, err := dbQueries.GetMeetingMessagesByID(ctx, internalMeetingID) // conn.Query(ctx, "SELECT internal_user_id, message_content, send_time FROM chat_messages WHERE internal_meeting_id = $1", internalMeetingID)
	if err != nil {
		return timeline, fmt.Errorf("Unable to obtaib meeting messages with Id %s: %v\n", internalMeetingID, err)
	}
	for _, message := range meetingMessages {
		var event = Event{}
		var userName string
		if !message.SendTime.Valid {
			fmt.Println("WARNING chat message has no sent time", message)
			continue
		}

		event.Time = message.SendTime.Time

		// Handle system message.
		if message.InternalUserID == "SYSTEM" {
			event.TextRepresentation = TranslateAdvanced("SystemSentMessage", map[string]string{"Message": message.MessageContent})
			timeline = append(timeline, event)
			continue
		}

		// we cannot use the same connection while the row is not closed.
		userName, err = dbQueries.GetUserNameById(ctx, message.InternalUserID)
		if err != nil {
			fmt.Println("error obtaining user information of chat message")
			fmt.Println(err)
		}
		event.TextRepresentation = TranslateAdvanced("ReportMessageEventRepresentation", map[string]string{"Username": userName, "Message": message.MessageContent})
		timeline = append(timeline, event)
	}
	return timeline, err
}

func FillMeetingUserEvents(ctx context.Context, internalMeetingID string, conn *pgx.Conn) (timeline []Event, err error) {
	// insert user events into the timeline
	row, err := conn.Query(ctx, "SELECT event_timestamp, internal_user_id, event_type FROM user_events WHERE internal_meeting_id = $1", internalMeetingID)
	if err != nil {
		return timeline, fmt.Errorf("Unable to find meeting with Id %s: %v\n", internalMeetingID, err)
	}
	for row.Next() {
		var event Event
		var userID string
		var userName string
		var eventType string

		con2, err := pgx.Connect(ctx, confGet("DB_CONNECTION_STRING"))
		if err != nil {
			fmt.Println("error connecting to the database at least twice. ->", err)
			return timeline, err
		}
		err = row.Scan(&event.Time, &userID, &eventType)
		if err != nil {
			fmt.Println(err)
		}

		err = con2.QueryRow(ctx, "SELECT name FROM users WHERE internal_user_id = $1", userID).Scan(&userName)
		if err != nil {
			fmt.Println("error obtaining user information of user event")
			return timeline, err
		}
		event.TextRepresentation = TranslateAdvanced(eventType, map[string]string{"Username": userName})
		timeline = append(timeline, event)
	}
	row.Close()
	return timeline, err
}

func FillMeetingParticipants(ctx context.Context, internalMeetingID string, conn *pgx.Conn) (error, []BBBEvents.User) {
	// Query participants and parse them
	var err error
	var participants []BBBEvents.User
	row, err := conn.Query(ctx, "SELECT DISTINCT internal_user_id FROM user_events WHERE internal_meeting_id = $1", internalMeetingID)
	if err != nil {
		return fmt.Errorf("Error occured when generating report for Id %s: %v\n", internalMeetingID, err), nil
	}
	var participantIds []string
	for row.Next() {
		var participantID string
		err = row.Scan(&participantID)
		if err != nil {
			fmt.Println(err)
		}
		participantIds = append(participantIds, participantID)
	}
	row.Close()

	for _, participantID := range participantIds {
		var user BBBEvents.User
		err = conn.QueryRow(ctx, "SELECT internal_user_id, external_user_id, name, role, is_guest FROM users WHERE internal_user_id = $1", participantID).Scan(&user.InternalUserID, &user.ExternalUserID, &user.Name, &user.Role, &user.Guest)
		if err != nil {
			fmt.Println("Error occurred adding user to participants list:", err)
			continue
		}
		participants = append(participants, user)
	}

	return err, participants
}

func FillMeetingDetails(details []Detail, meeting BBBEvents.Meeting) []Detail {
	// Load Details
	details = append(details, Detail{Translate("ReportMeetingDetailMeetingName"), meeting.Name})
	details = append(details, Detail{Translate("ReportMeetingDetailInternalMeetingID"), meeting.InternalMeetingID})
	details = append(details, Detail{Translate("ReportMeetingDetailBBBHostname"), meeting.BbbHostname})
	details = append(details, Detail{Translate("ReportMeetingDetailCreationDate"), meeting.CreateTime.Format("02.01.2006")})
	return details
}
