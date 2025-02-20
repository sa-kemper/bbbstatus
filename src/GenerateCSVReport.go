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
	"errors"
	"fmt"
	"github.com/jackc/pgx/v5"
	"github.com/nicksnyder/go-i18n/v2/i18n"
	"os"
	"sort"
	"strings"
	"time"
)

type CSVEvent struct {
	Time               time.Time
	FormattedTime      string
	User               string
	Action             string
	TextRepresentation string
}

func init() {
	messages := []i18n.Message{
		{ID: "CSVReportConfig-time", Other: "Time"},
		{ID: "CSVReportConfig-user", Other: "User"},
		{ID: "CSVReportConfig-action", Other: "Action"},
		{ID: "CSVReportConfig-textrepresentation", Other: "Text Representation"},
		{ID: "CSVReportActionChatted", Other: "Chatted"},
		{ID: "CSVReportAction-user-joined", Other: "Joined the meeting"},
		{ID: "CSVReportAction-user-left", Other: "Left the meeting"},
		{ID: "CSVReportAction-user-presenter-assigned", Other: "Assigned as presenter"},
		{ID: "CSVReportAction-user-presenter-unassigned", Other: "Unassigned as presenter"},
		{ID: "CSVReportAction-user-audio-voice-enabled", Other: "Joined the audio voice channel"},
		{ID: "CSVReportAction-user-audio-voice-disabled", Other: "Left the audio voice channel"},
		{ID: "CSVReportAction-user-audio-muted", Other: "Muted"},
		{ID: "CSVReportAction-user-audio-unmuted", Other: "Unmuted"},
		{ID: "CSVReportAction-user-cam-broadcast-start", Other: "Started a webcam broadcast"},
		{ID: "CSVReportAction-user-cam-broadcast-end", Other: "Ended a webcam broadcast"},
		{ID: "CSVReportAction-meeting-screenshare-started", Other: "Started a screenshare"},
		{ID: "CSVReportAction-meeting-screenshare-stopped", Other: "Stopped a screenshare"},
		{ID: "CSVReportAction-user-emoji-changed", Other: "Raised his hand"},
	}
	FrontendTextMessages = append(FrontendTextMessages, messages...)
}

func GenerateCSVReport(ctx context.Context, internalMeetingID string) ([]byte, error) {
	var result string
	var timeline []CSVEvent
	var polls *[]string
	conn, err := pgx.Connect(ctx, confGet("DB_CONNECTION_STRING"))
	if err != nil {
		if errors.Is(err, os.ErrDeadlineExceeded) {
			fmt.Println("FATAL Database timed out")
		}
		return nil, err
	}
	err = conn.Ping(ctx)
	if err != nil {
		fmt.Println("error occurred while connecting to database (GenerateCSVReport)")
		return nil, err
	}

	// Translate CSV Header
	splitResult := strings.Split(confGet("CSV_STRUCTURE"), ",")
	for _, value := range splitResult {
		value = strings.TrimSpace(value)
		value = strings.ReplaceAll(value, " ", "")
		value = strings.ToLower(value)
		result += Translate("CSVReportConfig-"+value) + ","
	}

	err, messageTimeline := fillCSVMessageEvents(ctx, internalMeetingID, conn)
	err, userEventTimeline := FillCSVUserEvents(ctx, internalMeetingID, conn)
	err, pollTimeline := FillCsvPollEvents(ctx, internalMeetingID, conn, polls)

	timeline = append(timeline, messageTimeline...)
	timeline = append(timeline, userEventTimeline...)
	timeline = append(timeline, pollTimeline...)

	for i, event := range timeline {
		timeline[i].FormattedTime = event.Time.Format(time.TimeOnly)
	}

	// sort the timeline
	sort.Slice(timeline, func(i, j int) bool {
		return timeline[i].Time.Before(timeline[j].Time)
	})

	CsvStructureConfig := strings.Split(confGet("CSV_STRUCTURE"), ",")
	for _, event := range timeline {
		result += event.ReturnCSVRow(CsvStructureConfig)
	}

	return []byte(result), nil
}

func (e CSVEvent) ReturnCSVRow(CsvStructureConfig []string) string {
	var result = "\n"
	for _, ConfiguredEventDetail := range CsvStructureConfig {
		switch ConfiguredEventDetail {
		case "text representation":
			result += fmt.Sprintf("%s,", e.TextRepresentation)
		case "action":
			result += fmt.Sprintf("%s,", e.Action)
		case "time":
			result += fmt.Sprintf("%s,", e.FormattedTime)
		case "user":
			result += fmt.Sprintf("%s,", e.User)

		}
	}
	return result
}

func fillCSVMessageEvents(ctx context.Context, internalMeetingID string, conn *pgx.Conn) (err error, timeline []CSVEvent) {
	// insert user messages into the timeline
	row, err := conn.Query(ctx, "SELECT internal_user_id, message_content, send_time FROM chat_messages WHERE internal_meeting_id = $1", internalMeetingID)
	if err != nil {
		return fmt.Errorf("Unable to find meeting with Id %s: %v\n", internalMeetingID, err), timeline
	}
	for row.Next() {
		var event CSVEvent
		var chatMessageContent string
		var chatMessageUserID string
		err = row.Scan(&chatMessageUserID, &chatMessageContent, &event.Time)
		if err != nil {
			fmt.Println(err)
		}

		// we cannot use the same connection while the row is not closed.
		conn2, err := pgx.Connect(ctx, confGet("DB_CONNECTION_STRING"))
		if err != nil {
			fmt.Println("error connecting to the database at least twice. ->", err)
		} else {
			err = conn2.QueryRow(ctx, "SELECT name FROM users WHERE internal_user_id = $1", chatMessageUserID).Scan(&event.User)
			if err != nil {
				fmt.Println("error obtaining user information of chat message")
				fmt.Println(err)
			}
			event.TextRepresentation = TranslateAdvanced("ReportMessageEventRepresentation", map[string]string{"Username": event.User, "Message": chatMessageContent})
			event.Action = Translate("CSVReportActionChatted")
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

func FillCSVUserEvents(ctx context.Context, internalMeetingID string, conn *pgx.Conn) (err error, timeline []CSVEvent) {
	// insert user events into the timeline
	row, err := conn.Query(ctx, "SELECT event_timestamp, internal_user_id, event_type FROM user_events WHERE internal_meeting_id = $1", internalMeetingID)
	if err != nil {
		return fmt.Errorf("Unable to find meeting with Id %s: %v\n", internalMeetingID, err), timeline
	}
	for row.Next() {
		var event CSVEvent
		var userID string
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

		err = con2.QueryRow(ctx, "SELECT name FROM users WHERE internal_user_id = $1", userID).Scan(&event.User)
		if err != nil {
			fmt.Println("error obtaining user information of user event")
			return err, timeline
		}
		event.TextRepresentation = TranslateAdvanced(eventType, map[string]string{"Username": event.User})
		event.Action = Translate("CSVReportAction-" + eventType)
		timeline = append(timeline, event)
	}
	row.Close()
	return err, timeline
}

func FillCsvPollEvents(ctx context.Context, internalMeetingID string, conn *pgx.Conn, polls *[]string) (err error, timeline []CSVEvent) {
	// insert the polls into the timeline
	row, err := conn.Query(ctx, "SELECT poll_id, internal_user_id, question, answers, created_at FROM polls WHERE internal_meeting_id = $1", internalMeetingID)
	if err != nil {
		return fmt.Errorf("Unable to find meeting with Id %s: %v\n", internalMeetingID, err), timeline
	}
	for row.Next() {
		var event CSVEvent
		var pollId string
		var userID string
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
		err = conn2.QueryRow(ctx, "SELECT name FROM users WHERE internal_user_id = $1", userID).Scan(&event.User)
		if err != nil {
			fmt.Println("error obtaining user information of chat message")
			fmt.Println(err)
		}

		if question != "" { // make a suitable question replacement in case the user drops the question content altogether.
			question = string([]byte(pollId)[0:5])
		}

		event.TextRepresentation = TranslateAdvanced("ReportPollStartedEventRepresentation", map[string]string{"Username": event.User, "PollQuestion": question, "PollOptions": answersTextRepresentation})
		event.Action = Translate("CSVReportActionPollStarted")
		timeline = append(timeline, event)
		//goland:noinspection ALL
		conn2.Close(ctx)

	}
	row.Close()
	return err, timeline
}

func FillCsvPollResponses(ctx context.Context, internalMeetingID string, polls *[]string, conn *pgx.Conn) (err error, timeline []CSVEvent) {
	// insert the poll Answers into the timeline
	for _, pollID := range *polls {
		row, err := conn.Query(ctx, "SELECT internal_user_id, answer_ids, response_time FROM poll_responses WHERE poll_id = $1", pollID)
		if err != nil {
			return fmt.Errorf("Unable to find meeting with Id %s: %v\n", internalMeetingID, err), timeline
		}
		for row.Next() {
			var event CSVEvent
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
			event.TextRepresentation = TranslateAdvanced("ReportPollResponseEventRepresentation", map[string]string{"username": user.Name, "pollQuestion": poll.Question, "pollAnswer": poll.Answers[poll.AnswerIds[0]].Key})
			event.Action = Translate("CSVReportActionPollResponse")
			event.User = user.Name
			timeline = append(timeline, event)
			//goland:noinspection ALL
			conn2.Close(ctx)

		}
		row.Close()
	}
	return err, timeline
}
