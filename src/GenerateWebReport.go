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
	"BbbStatus/internal/BBBEvents"
	"context"
	"encoding/json"
	"fmt"
	"github.com/jackc/pgx/v5"
	"github.com/nicksnyder/go-i18n/v2/i18n"
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

	// Connect to the db using
	conn, err := pgx.Connect(ctx, confGet("DB_CONNECTION_STRING"))
	if err != nil {
		return Report{}, fmt.Errorf("Unable to connect to database: %v\n", err)
	}
	//goland:noinspection ALL
	defer conn.Close(ctx)

	// Query and parse meeting using the row.next and row.scan methode of pgx
	err = conn.QueryRow(ctx, "SELECT * FROM meetings WHERE internal_meeting_id = $1 LIMIT 1", internalMeetingID).Scan(&meeting.InternalMeetingID, &meeting.ExternalMeetingID, &meeting.Name, &meeting.IsBreakout, &meeting.ParentID, &meeting.Duration, &meeting.CreateTime, &meeting.ModeratorPass, &meeting.ViewerPass, &meeting.Record, &meeting.VoiceConf, &meeting.DialNumber, &meeting.MaxUsers, &meeting.Metadata, &meeting.BbbHostname, &meeting.Active)
	if err != nil {
		return Report{}, fmt.Errorf("Unable to find meeting with Id %s: %v\n", internalMeetingID, err)
	}

	details = FillMeetingDetails(details, meeting)

	err, participants = FillMeetingParticipants(ctx, internalMeetingID, conn)
	if err != nil {
		return Report{}, err
	}

	err, userEvents := FillMeetingUserEvents(ctx, internalMeetingID, conn)
	if err != nil {
		return Report{}, err
	}
	timeline = append(timeline, userEvents...)

	err, messageTimeline := FillMeetingMessageEvents(ctx, internalMeetingID, conn)
	if err != nil {
		return Report{}, err
	}
	timeline = append(timeline, messageTimeline...)

	err, pollTimeline := FillMeetingPollEvents(ctx, internalMeetingID, conn, &polls)
	if err != nil {
		return Report{}, err
	}
	timeline = append(timeline, pollTimeline...)

	err, pollResponseTimeline := FillMeetingPollResponses(ctx, internalMeetingID, &polls, conn)
	if err != nil {
		return Report{}, err
	}
	timeline = append(timeline, pollResponseTimeline...)

	for i, event := range timeline {
		timeline[i].FormattedTime = event.Time.Format(time.TimeOnly + " 02.01.2006")
	}

	// sort the timeline
	sort.Slice(timeline, func(i, j int) bool {
		return timeline[i].Time.Before(timeline[j].Time)
	})

	return Report{Details: details, Participants: participants, Timeline: timeline}, nil
}

func FillMeetingPollResponses(ctx context.Context, internalMeetingID string, polls *[]string, conn *pgx.Conn) (err error, timeline []Event) {
	// insert the poll Answers into the timeline
	for _, pollID := range *polls {
		row, err := conn.Query(ctx, "SELECT internal_user_id, answer_ids, response_time FROM poll_responses WHERE poll_id = $1", pollID)
		if err != nil {
			return fmt.Errorf("Unable to find meeting with Id %s: %v\n", internalMeetingID, err), timeline
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
				return err, timeline
			}

			// we cannot use the same connection while the row is not closed.
			conn2, err := pgx.Connect(ctx, confGet("DB_CONNECTION_STRING"))
			if err != nil {
				fmt.Println("error connecting to the database at least twice. ->", err)
				return err, timeline
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
	return err, timeline
}

func FillMeetingPollEvents(ctx context.Context, internalMeetingID string, conn *pgx.Conn, polls *[]string) (err error, timeline []Event) {
	// insert the polls into the timeline
	row, err := conn.Query(ctx, "SELECT poll_id, internal_user_id, question, answers, created_at FROM polls WHERE internal_meeting_id = $1", internalMeetingID)
	if err != nil {
		return fmt.Errorf("Unable to find meeting with Id %s: %v\n", internalMeetingID, err), timeline
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
			return err, timeline
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
	return err, timeline
}

func FillMeetingMessageEvents(ctx context.Context, internalMeetingID string, conn *pgx.Conn) (err error, timeline []Event) {
	// insert user messages into the timeline
	row, err := conn.Query(ctx, "SELECT internal_user_id, message_content, send_time FROM chat_messages WHERE internal_meeting_id = $1", internalMeetingID)
	if err != nil {
		return fmt.Errorf("Unable to find meeting with Id %s: %v\n", internalMeetingID, err), timeline
	}
	for row.Next() {
		var event Event
		var chatMessageContent string
		var chatMessageUserID string
		var userName string
		err = row.Scan(&chatMessageUserID, &chatMessageContent, &event.Time)
		if err != nil {
			fmt.Println(err)
		}

		// we cannot use the same connection while the row is not closed.
		conn2, err := pgx.Connect(ctx, confGet("DB_CONNECTION_STRING"))
		if err != nil {
			fmt.Println("error connecting to the database at least twice. ->", err)
		} else {
			err = conn2.QueryRow(ctx, "SELECT name FROM users WHERE internal_user_id = $1", chatMessageUserID).Scan(&userName)
			if err != nil {
				fmt.Println("error obtaining user information of chat message")
				fmt.Println(err)
			}
			event.TextRepresentation = TranslateAdvanced("ReportMessageEventRepresentation", map[string]string{"Username": userName, "Message": chatMessageContent})
			timeline = append(timeline, event)
			err = conn2.Close(ctx)
			if err != nil {
				fmt.Println("error closing the database connection when getting additional user information")
				return err, timeline
			}
		}
	}
	row.Close()
	return err, timeline
}

func FillMeetingUserEvents(ctx context.Context, internalMeetingID string, conn *pgx.Conn) (err error, timeline []Event) {
	// insert user events into the timeline
	row, err := conn.Query(ctx, "SELECT event_timestamp, internal_user_id, event_type FROM user_events WHERE internal_meeting_id = $1", internalMeetingID)
	if err != nil {
		return fmt.Errorf("Unable to find meeting with Id %s: %v\n", internalMeetingID, err), timeline
	}
	for row.Next() {
		var event Event
		var userID string
		var userName string
		var eventType string

		con2, err := pgx.Connect(ctx, confGet("DB_CONNECTION_STRING"))
		if err != nil {
			fmt.Println("error connecting to the database at least twice. ->", err)
			return err, timeline
		}
		err = row.Scan(&event.Time, &userID, &eventType)
		if err != nil {
			fmt.Println(err)
		}

		err = con2.QueryRow(ctx, "SELECT name FROM users WHERE internal_user_id = $1", userID).Scan(&userName)
		if err != nil {
			fmt.Println("error obtaining user information of user event")
			return err, timeline
		}
		event.TextRepresentation = TranslateAdvanced(eventType, map[string]string{"Username": userName})
		timeline = append(timeline, event)
	}
	row.Close()
	return err, timeline
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
		err = conn.QueryRow(ctx, "SELECT * FROM users WHERE internal_user_id = $1", participantID).Scan(&user.InternalUserID, &user.ExternalUserID, &user.Name, &user.Role, &user.Guest)
		if err != nil {
			fmt.Println(err)
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
