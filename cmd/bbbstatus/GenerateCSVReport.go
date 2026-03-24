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
	db "bbbstatus/internal/database"
	"bbbstatus/locales"
	"bbbstatus/pkg/BBBEvents"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"slices"
	"sort"
	"strings"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/nicksnyder/go-i18n/v2/i18n"
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
		{ID: "CSVReportConfig-time", Other: "time"},
		{ID: "CSVReportConfig-user", Other: "user"},
		{ID: "CSVReportConfig-action", Other: "action"},
		{ID: "CSVReportConfig-textrepresentation", Other: "text Representation"},
		{ID: "CSVReportActionChatted", Other: "chatted"},
		{ID: "CSVReportAction-user-joined", Other: "joined the meeting"},
		{ID: "CSVReportAction-user-left", Other: "left the meeting"},
		{ID: "CSVReportAction-user-presenter-assigned", Other: "assigned as presenter"},
		{ID: "CSVReportAction-user-presenter-unassigned", Other: "unassigned as presenter"},
		{ID: "CSVReportAction-user-audio-voice-enabled", Other: "joined the audio voice channel"},
		{ID: "CSVReportAction-user-audio-voice-disabled", Other: "left the audio voice channel"},
		{ID: "CSVReportAction-user-audio-muted", Other: "muted"},
		{ID: "CSVReportAction-user-audio-unmuted", Other: "unmuted"},
		{ID: "CSVReportAction-user-cam-broadcast-start", Other: "started a webcam broadcast"},
		{ID: "CSVReportAction-user-cam-broadcast-end", Other: "ended a webcam broadcast"},
		{ID: "CSVReportAction-meeting-screenshare-started", Other: "started screensharing"},
		{ID: "CSVReportAction-meeting-screenshare-stopped", Other: "stopped a screensharing"},
		{ID: "CSVReportAction-user-emoji-changed", Other: "raised his hand"},
		{ID: "SystemMessage", Other: "System message"},
		{ID: "CSVMeetingEvent-meeting-created", Other: "meeting created"},
		{ID: "CSVMeetingEvent-meeting-ended", Other: "meeting ended"},
		{ID: "CSVMeetingEvent-meeting-recording-started", Other: "the meeting recording started"},
		{ID: "CSVMeetingEvent-meeting-recording-stopped", Other: "the meeting recording stopped"},
		{ID: "CSVMeetingAction-meeting-recording-started", Other: "recording start"},
		{ID: "CSVMeetingAction-meeting-recording-stopped", Other: "recording stopped"},
		{ID: "CSVMeetingAction-meeting-created", Other: "meeting created"},
		{ID: "CSVMeetingAction-meeting-ended", Other: "meeting ended"},
	}
	FrontendTextMessages = append(FrontendTextMessages, messages...)
}

func GenerateCSVReport(ctx context.Context, internalMeetingID string, filteredUserIds *[]string) ([]byte, error) {
	var result string
	var timeline []CSVEvent
	var polls []string
	conn, err := pgx.Connect(ctx, confGet("DB_CONNECTION_STRING"))
	if err != nil {
		if errors.Is(err, os.ErrDeadlineExceeded) {
			fmt.Println("FATAL Database timed out")
		}
		return nil, err
	}
	defer conn.Close(ctx)

	err = conn.Ping(ctx)
	if err != nil {
		fmt.Println("error occurred while connecting to database (GenerateCSVReport)")
		return nil, err
	}

	dbQueries := db.New(conn)

	// Translate CSV Header
	splitResult := strings.Split(confGet("CSV_STRUCTURE"), ",")
	for _, value := range splitResult {
		value = strings.TrimSpace(value)
		value = strings.ReplaceAll(value, " ", "")
		value = strings.ToLower(value)
		result += locales.TranslateFromCTX(ctx, "CSVReportConfig-"+value) + ","
	}

	messageTimeline, err := fillCSVMessageEvents(ctx, internalMeetingID, dbQueries, filteredUserIds)
	if err != nil {
		fmt.Println("error occurred while filling message timeline (GenerateCSVReport)", err.Error())
		return nil, err
	}

	userEventTimeline, err := FillCSVUserEvents(ctx, internalMeetingID, dbQueries, filteredUserIds)
	if err != nil {
		fmt.Println("error occurred while filling user event timeline (GenerateCSVReport)", err.Error())
		return nil, err
	}

	pollTimeline, err := FillCsvPollEvents(ctx, internalMeetingID, conn, &polls)
	if err != nil {
		fmt.Println("error occurred while filling poll timeline (GenerateCSVReport)", err.Error())
		return nil, err
	}

	pollResponseTimeline, err := FillCsvPollResponses(ctx, internalMeetingID, &polls, conn)
	if err != nil {
		fmt.Println("error occurred while filling poll responses (GenerateCSVReport)", err.Error())
		return nil, err
	}

	meetingEventTimeline, err := FillCsvMeetingEvents(ctx, internalMeetingID, dbQueries)
	if err != nil {
		fmt.Println("error occurred while filling meeting event timeline (GenerateCSVReport)", err.Error())
		return nil, err
	}

	timeline = append(timeline, messageTimeline...)
	timeline = append(timeline, userEventTimeline...)
	timeline = append(timeline, pollTimeline...)
	timeline = append(timeline, pollResponseTimeline...)
	timeline = append(timeline, meetingEventTimeline...)

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

func FillCsvMeetingEvents(ctx context.Context, internalMeetingId string, dbQueries *db.Queries) (timeline []CSVEvent, err error) {
	meetingEvents, err := dbQueries.GetMeetingEventsByInternalMeetingID(ctx, internalMeetingId)
	if err != nil {
		return nil, err
	}
	for _, dbEvent := range meetingEvents {
		var event = CSVEvent{Time: dbEvent.EventTimestamp.Time, User: "SYSTEM", TextRepresentation: locales.TranslateFromCTX(ctx, "CSVMeetingAction-"+dbEvent.EventType)}
		timeline = append(timeline, event)
	}

	return timeline, nil
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

func fillCSVMessageEvents(ctx context.Context, internalMeetingID string, dbQueries *db.Queries, filteredForUserIds *[]string) (timeline []CSVEvent, err error) {
	// insert user messages into the timeline
	messages, err := dbQueries.GetMeetingMessagesByID(ctx, internalMeetingID)
	if err != nil {
		return timeline, fmt.Errorf("unable to find meeting with Id %s: %v", internalMeetingID, err)
	}
	for _, message := range messages {
		var event = CSVEvent{Time: message.SendTime.Time}
		var chatMessageContent = message.MessageContent
		var chatMessageUserID = message.InternalUserID

		if !message.SendTime.Valid {
			fmt.Println("WARNING: message send time is not valid, this is likely a bug")
		}

		// Handle system messages
		if chatMessageUserID == "SYSTEM" {
			event.Action = locales.TranslateFromCTX(ctx, "SystemMessage")
			event.TextRepresentation = locales.TranslateAdvanced(ctx, "SystemSentMessage", map[string]string{"Message": chatMessageContent})
			event.FormattedTime = event.Time.Format(time.TimeOnly)
			event.User = chatMessageUserID
			timeline = append(timeline, event)
			continue
		}

		if filteredForUserIds != nil {
			if !slices.Contains(*filteredForUserIds, message.InternalUserID) {
				continue
			}
		}

		eventUser, err := dbQueries.GetUserById(ctx, message.InternalUserID)
		if err != nil {
			fmt.Println("error obtaining user information of chat message")
			fmt.Println(err)
			event.User = chatMessageUserID
		}
		event.User = eventUser.Name
		event.TextRepresentation = locales.TranslateAdvanced(ctx, "ReportMessageEventRepresentation", map[string]string{"Username": event.User, "Message": chatMessageContent})
		event.Action = locales.TranslateFromCTX(ctx, "CSVReportActionChatted")
		timeline = append(timeline, event)
	}
	return timeline, err
}

func FillCSVUserEvents(ctx context.Context, internalMeetingID string, dbQueries *db.Queries, filteredForUserIds *[]string) (timeline []CSVEvent, err error) {
	// insert user events into the timeline
	userEvents, err := dbQueries.GetUserEventsByMeetingID(ctx, internalMeetingID)
	if err != nil {
		return timeline, fmt.Errorf("unable to find meeting with Id %s: %v", internalMeetingID, err)
	}
	for _, dbEvent := range userEvents {
		if filteredForUserIds != nil {
			if !slices.Contains(*filteredForUserIds, dbEvent.InternalUserID) {
				continue
			}
		}
		var event = CSVEvent{Time: dbEvent.EventTimestamp.Time}
		eventUser, err := dbQueries.GetUserById(ctx, dbEvent.InternalUserID)
		if err != nil {
			fmt.Println("error obtaining user information of user event")
			return timeline, err
		}
		event.User = eventUser.Name
		event.TextRepresentation = locales.TranslateAdvanced(ctx, dbEvent.EventType, map[string]string{"Username": event.User})
		event.Action = locales.TranslateFromCTX(ctx, "CSVReportAction-"+dbEvent.EventType)
		timeline = append(timeline, event)
	}
	return timeline, err
}

func FillCsvPollEvents(ctx context.Context, internalMeetingID string, conn *pgx.Conn, polls *[]string) (timeline []CSVEvent, err error) {
	// insert the polls into the timeline
	row, err := conn.Query(ctx, "SELECT poll_id, internal_user_id, question, answers, created_at FROM polls WHERE internal_meeting_id = $1", internalMeetingID)
	if err != nil {
		return timeline, fmt.Errorf("unable to find meeting with Id %s: %v", internalMeetingID, err)
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
			return timeline, err
		}
		err = conn2.QueryRow(ctx, "SELECT name FROM users WHERE internal_user_id = $1", userID).Scan(&event.User)
		if err != nil {
			fmt.Println("error obtaining user information of chat message")
			fmt.Println(err)
		}

		if question != "" { // make a suitable question replacement in case the user drops the question content altogether.
			question = string([]byte(pollId)[0:5])
		}

		event.TextRepresentation = locales.TranslateAdvanced(ctx, "ReportPollStartedEventRepresentation", map[string]string{"Username": event.User, "PollQuestion": question, "PollOptions": answersTextRepresentation})
		event.Action = locales.TranslateFromCTX(ctx, "CSVReportActionPollStarted")
		timeline = append(timeline, event)
		//goland:noinspection ALL
		conn2.Close(ctx)

	}
	row.Close()
	return timeline, err
}

func FillCsvPollResponses(ctx context.Context, internalMeetingID string, polls *[]string, conn *pgx.Conn) (timeline []CSVEvent, err error) {
	// insert the poll Answers into the timeline
	for _, pollID := range *polls {
		row, err := conn.Query(ctx, "SELECT internal_user_id, answer_ids, response_time FROM poll_responses WHERE poll_id = $1", pollID)
		if err != nil {
			return timeline, fmt.Errorf("unable to find meeting with Id %s: %v", internalMeetingID, err)
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
				return timeline, err
			}

			if poll.Question == "" { // use a shortened version of the poll id if the question is empty.
				poll.Question = string([]byte(pollID)[0:5])
			}
			event.TextRepresentation = locales.TranslateAdvanced(ctx, "ReportPollResponseEventRepresentation", map[string]string{"username": user.Name, "pollQuestion": poll.Question, "pollAnswer": poll.Answers[poll.AnswerIds[0]].Key})
			event.Action = locales.TranslateFromCTX(ctx, "CSVReportActionPollResponse")
			event.User = user.Name
			timeline = append(timeline, event)
			//goland:noinspection ALL
			conn2.Close(ctx)

		}
		row.Close()
	}
	return timeline, err
}
