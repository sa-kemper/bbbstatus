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
	"bbbstatus/internal/config"
	db "bbbstatus/internal/database"
	"bbbstatus/locales"
	"bbbstatus/pkg/BBBAPI"
	"bbbstatus/pkg/BBBEvents"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"slices"
	"sort"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/nicksnyder/go-i18n/v2/i18n"
)

type Detail struct {
	Name  string
	Value string
}

type Event struct {
	Time               time.Time
	FormattedTime      string
	EventType          BBBEvents.BBBEventType
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
		{ID: "EventTimelineHeader", Other: "events timeline"},
		{ID: "ReportMeetingDetailMeetingName", Other: "meeting name"},
		{ID: "ReportHeader", Other: "meeting report"},
		{ID: "ReportMeetingDetailsHeader", Other: "meeting details"},
		{ID: "ReportParticipantsHeader", Other: "participants"},
		{ID: "ReportPollResponseEventRepresentation", Other: "User '{{.Username}}' responded to a poll {{.PollQuestion}} with {{.PollAnswer}}"},
		{ID: "ReportPollStartedEventRepresentation", Other: "User '{{.Username}}' started a poll '{{.PollQuestion}}', with options: {{.PollOptions}}"},
		{ID: "ReportMessageEventRepresentation", Other: "User '{{.Username}}' sent a message: {{.Message}}"},
		{ID: "ReportPrintReportButton", Other: "print report"},
		{ID: "ReportOpenInExcelButton", Other: "open in Excel"},
		{ID: "ReportMeetingDetailInternalMeetingID", Other: "internal meeting ID"},
		{ID: "ReportMeetingDetailBBBHostname", Other: "BBB hostname"},
		{ID: "ReportMeetingDetailCreationDate", Other: "creation date"},
		{ID: "meetingListHeader", Other: "meeting list"},
		{ID: "BackToMeetingsButton", Other: "back to meetings"},
		{ID: "RecordingsHeader", Other: "recordings"},
		{ID: "SystemSentMessage", Other: "System Sent the message: '{{.Message}}'"},
		{ID: "MeetingEvent-meeting-created", Other: "meeting created"},
		{ID: "MeetingEvent-meeting-ended", Other: "meeting ended"},
		{ID: "MeetingEvent-meeting-recording-started", Other: "Meeting Recording Started"},
		{ID: "MeetingEvent-meeting-recording-stopped", Other: "Meeting Recording Stoped"},
	}
	FrontendTextMessages = append(FrontendTextMessages, msgs...)
	for _, m := range BBBEvents.UserEventTextRepresentation { // Add user events text representation to the language strings.
		FrontendTextMessages = append(FrontendTextMessages, m)
	}
}

func GenerateWebReport(ctx context.Context, conf *config.ConfigurationStruct, internalMeetingID string, filteredForUserIds *[]string) (Report, error) {
	var details = make([]Detail, 0)
	var participants []BBBEvents.User
	var timeline []Event
	var polls []string
	var meetingServerAPI BBBAPI.API
	var recordings []BBBAPI.Recording

	// Connect to the db using
	conn, err := pgx.Connect(ctx, conf.DatabaseConfig.DatabaseConnectionString)
	if err != nil {
		return Report{}, fmt.Errorf("unable to connect to database: %v", err)
	}
	//goland:noinspection ALL
	defer conn.Close(ctx)

	dbQueries := db.New(conn)

	// Query and parse meeting using the row.next and row.scan methode of pgx
	meeting, err := dbQueries.GetMeetingById(ctx, internalMeetingID)
	if err != nil {
		return Report{}, fmt.Errorf("unable to find meeting with Id %s: %v", internalMeetingID, err)
	}

	ScaleliteServers := conf.FindScaleliteServers("")
	servers := conf.FindBBBServers(meeting.Bbbhostname)
	var server config.BbbServer

	if len(servers) < 1 {
		return Report{}, fmt.Errorf("unable to find server with hostname %s", meeting.Bbbhostname)
	}

	server = servers[0]
	if server.Hostname == meeting.Bbbhostname {
		if server.APITimeout != 0 {
			meetingServerAPI = BBBAPI.API{Hostname: server.Hostname, Port: server.ApiPort, SharedSecret: server.SharedSecret, Timeout: new(time.Duration(server.APITimeout) * time.Second)}
		} else {
			meetingServerAPI = BBBAPI.API{Hostname: server.Hostname, Port: server.ApiPort, SharedSecret: server.SharedSecret}
		}
	} else {
		//fmt.Println("DEBUG API SERVER FINDING: '" + server.Hostname + "' != '" + meeting.BbbHostname + "'")
	}

	details = append(details, Detail{locales.TranslateFromCTX(ctx, "ReportMeetingDetailMeetingName"), meeting.Name})
	details = append(details, Detail{locales.TranslateFromCTX(ctx, "ReportMeetingDetailInternalMeetingID"), meeting.InternalMeetingID})
	details = append(details, Detail{locales.TranslateFromCTX(ctx, "ReportMeetingDetailBBBHostname"), meeting.Bbbhostname})
	details = append(details, Detail{locales.TranslateFromCTX(ctx, "ReportMeetingDetailCreationDate"), meeting.CreateTime.Time.Format("02.01.2006")})

	participants, err = FillMeetingParticipants(ctx, dbQueries, internalMeetingID, meeting)
	if err != nil {
		return Report{}, err
	}

	// Get recordings for this meeting
	if len(ScaleliteServers) > 0 {
		for _, scaleServer := range ScaleliteServers {
			meetingServerAPI = BBBAPI.API{Hostname: scaleServer.Hostname, Port: scaleServer.ApiPort, SharedSecret: scaleServer.SharedSecret, Timeout: new(time.Duration(scaleServer.APITimeout) * time.Second)}
			getRecordingsResponse, err := meetingServerAPI.GetRecordings(ctx, BBBAPI.GetRecordingsParameters{MeetingID: &meeting.ExternalMeetingID}, meeting)
			if err != nil {
				if errors.Is(err, os.ErrDeadlineExceeded) {
					fmt.Println("FATAL: BBB API Timeout for host:", meetingServerAPI.Hostname)
				}
				//fmt.Println("DEBUG: Error occurred whilst geting recordings. ", err)
				return Report{}, err
			}
			recordings = append(recordings, getRecordingsResponse.Recording...)
		}
	} else {
		if meetingServerAPI.Hostname != "" && meetingServerAPI.SharedSecret != "" {
			//fmt.Println("DEBUG: meetingServerAPI is valid enough ^^ ", meetingServerAPI)
			getRecordingsResponse, err := meetingServerAPI.GetRecordings(ctx, BBBAPI.GetRecordingsParameters{MeetingID: &meeting.ExternalMeetingID}, meeting)
			if err != nil {
				if errors.Is(err, os.ErrDeadlineExceeded) {
					fmt.Println("FATAL: BBB API Timeout for host:", meetingServerAPI.Hostname)
				}
				//fmt.Println("DEBUG: Error occurred whilst getting recordings. ", err)
				return Report{}, err
			}
			recordings = getRecordingsResponse.Recording
			//fmt.Println("DEBUG: recordings: ", recordings)
		}
	}

	userEvents, err := FillMeetingUserEvents(ctx, internalMeetingID, dbQueries, filteredForUserIds)
	if err != nil {
		return Report{}, err
	}
	timeline = append(timeline, userEvents...)

	messageTimeline, err := FillMeetingMessageEvents(ctx, internalMeetingID, dbQueries, filteredForUserIds)
	if err != nil {
		return Report{}, err
	}
	timeline = append(timeline, messageTimeline...)

	pollTimeline, err := FillMeetingPollEvents(ctx, conf, internalMeetingID, conn, &polls)
	if err != nil {
		return Report{}, err
	}
	timeline = append(timeline, pollTimeline...)

	pollResponseTimeline, err := FillMeetingPollResponses(ctx, conf, internalMeetingID, &polls, conn)
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
			TextRepresentation: locales.TranslateFromCTX(ctx, "MeetingEvent-"+event.EventType),
			EventType:          BBBEvents.BBBEventType(event.EventType),
		})
	}
	return timeline, nil
}

func FillMeetingPollResponses(ctx context.Context, conf *config.ConfigurationStruct, internalMeetingID string, polls *[]string, conn *pgx.Conn) (timeline []Event, err error) {
	// insert the poll Answers into the timeline
	for _, pollID := range *polls {
		row, err := conn.Query(ctx, "SELECT internal_user_id, answer_ids, response_time FROM poll_responses WHERE poll_id = $1", pollID)
		if err != nil {
			return timeline, fmt.Errorf("unable to find meeting with Id %s: %v", internalMeetingID, err)
		}
		for row.Next() {
			var event Event
			var poll = BBBEvents.Poll{ID: pollID}
			var user = BBBEvents.User{}
			var answerJSON string
			var answersJSON string
			var dbQueries = db.New(conn)
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
			conn2, err := pgx.Connect(ctx, conf.DatabaseConfig.DatabaseConnectionString)
			if err != nil {
				fmt.Println("error connecting to the database at least twice. ->", err)
				return timeline, err
			}
			pollUser, err := dbQueries.GetUserById(ctx, user.InternalUserID)
			if err != nil {
				fmt.Println("error obtaining user information of poll answer")
				fmt.Println(err)
				user.Name = user.InternalUserID
			}
			user.Name = pollUser.Name

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
			event.TextRepresentation = locales.TranslateAdvanced(ctx, "ReportPollResponseEventRepresentation", map[string]string{"username": user.Name, "pollQuestion": poll.Question, "pollAnswer": poll.Answers[poll.AnswerIds[0]].Key})
			event.EventType = BBBEvents.EventPollResponded
			timeline = append(timeline, event)
			//goland:noinspection ALL
			conn2.Close(ctx)

		}
		row.Close()
	}
	return timeline, err
}

func FillMeetingPollEvents(ctx context.Context, conf *config.ConfigurationStruct, internalMeetingID string, conn *pgx.Conn, polls *[]string) (timeline []Event, err error) {
	// insert the polls into the timeline
	row, err := conn.Query(ctx, "SELECT poll_id, internal_user_id, question, answers, created_at FROM polls WHERE internal_meeting_id = $1", internalMeetingID)
	if err != nil {
		return timeline, fmt.Errorf("unable to find meeting with Id %s: %v", internalMeetingID, err)
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
		conn2, err := pgx.Connect(ctx, conf.DatabaseConfig.DatabaseConnectionString)
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

		if ctx.Value(Gdpr("gdpr")).(bool) {
			event.TextRepresentation = locales.TranslateAdvanced(ctx, "ReportPollStartedEventRepresentation", map[string]string{"Username": userName, "PollQuestion": question, "PollOptions": answersTextRepresentation})
		} else {
			event.TextRepresentation = locales.TranslateAdvanced(ctx, "ReportPollStartedEventRepresentation", map[string]string{"Username": userName, "PollQuestion": question, "PollOptions": answersTextRepresentation})
		}
		event.EventType = BBBEvents.EventPollStarted
		timeline = append(timeline, event)
		//goland:noinspection ALL
		conn2.Close(ctx)

	}
	row.Close()
	return timeline, err
}

func FillMeetingMessageEvents(ctx context.Context, internalMeetingID string, dbQueries *db.Queries, filteredForUserIds *[]string) (timeline []Event, err error) {
	// insert user messages into the timeline
	meetingMessages, err := dbQueries.GetMeetingMessagesByID(ctx, internalMeetingID)
	if err != nil {
		return timeline, fmt.Errorf("unable to obtain meeting messages with Id %s: %v", internalMeetingID, err)
	}
	for _, message := range meetingMessages {
		var event = Event{}
		if !message.SendTime.Valid {
			fmt.Println("WARNING chat message has no sent time", message)
			continue
		}

		event.Time = message.SendTime.Time

		// Handle system message.
		if message.InternalUserID == "SYSTEM" {
			event.TextRepresentation = locales.TranslateAdvanced(ctx, "SystemSentMessage", map[string]string{"Message": message.MessageContent})
			timeline = append(timeline, event)
			continue
		}

		if filteredForUserIds != nil {
			if !slices.Contains(*filteredForUserIds, message.InternalUserID) {
				continue
			}
		}

		// we cannot use the same connection while the row is not closed.
		messageUser, err := dbQueries.GetUserById(ctx, message.InternalUserID)
		if err != nil {
			fmt.Println("error obtaining user information of chat message")
			fmt.Println(err)
		}
		if ctx.Value(Gdpr("gdpr")).(bool) {
			event.TextRepresentation = locales.TranslateAdvanced(ctx, "ReportMessageEventRepresentation", map[string]string{"Username": messageUser.GdprName, "Message": message.MessageContent})
		} else {
			event.TextRepresentation = locales.TranslateAdvanced(ctx, "ReportMessageEventRepresentation", map[string]string{"Username": messageUser.Name, "Message": message.MessageContent})
		}
		event.EventType = BBBEvents.EventChatGroupMessageSent
		timeline = append(timeline, event)
	}
	return timeline, err
}

func FillMeetingUserEvents(ctx context.Context, internalMeetingID string, dbQueries *db.Queries, filteredForUserIds *[]string) (timeline []Event, err error) {
	// insert user events into the timeline
	meetingEvents, err := dbQueries.GetUserEventsByMeetingID(ctx, internalMeetingID)
	if err != nil {
		return timeline, fmt.Errorf("unable to find meeting with Id %s: %v", internalMeetingID, err)
	}

	for _, userEvent := range meetingEvents { // convert each row in the db to a Event object so it can be used by the frontend template
		if filteredForUserIds != nil {
			if !slices.Contains(*filteredForUserIds, userEvent.InternalUserID) {
				continue
			}
		}

		eventUser, err := dbQueries.GetUserById(ctx, userEvent.InternalUserID) // obtain username for the text representation
		if err != nil {
			fmt.Println("WARNING obtaining user information of user event", userEvent)
			continue
		}

		if !userEvent.EventTimestamp.Valid { // handle unlikely but possible
			fmt.Println("WARNING user event has no sent time", userEvent)
			continue
		}
		var event Event
		if ctx.Value(Gdpr("gdpr")).(bool) {
			event = Event{Time: userEvent.EventTimestamp.Time, TextRepresentation: locales.TranslateAdvanced(ctx, userEvent.EventType, map[string]string{"Username": eventUser.GdprName})}
		} else {
			event = Event{Time: userEvent.EventTimestamp.Time, TextRepresentation: locales.TranslateAdvanced(ctx, userEvent.EventType, map[string]string{"Username": eventUser.Name})}
		}
		event.EventType = BBBEvents.BBBEventType(userEvent.EventType)
		timeline = append(timeline, event)
	}
	return timeline, err
}

func FillMeetingParticipants(ctx context.Context, dbQueries *db.Queries, internalMeetingID string, meeting db.Meeting) (participants []BBBEvents.User, err error) {
	// Query participants and parse them
	meetingUsers, err := dbQueries.GetUserIDsFromMeetingByMeetingID(ctx, internalMeetingID)
	if err != nil {
		return nil, fmt.Errorf("error occurred when generating report for Id %s: %v", internalMeetingID, err)
	}
	for _, participantID := range meetingUsers {
		dbUser, err := dbQueries.GetUserById(ctx, participantID)
		if err != nil {
			fmt.Println("Error occurred adding user to participants list:", err)
			continue
		}
		var states []db.GetUserEventsStateRow
		states, err = dbQueries.GetUserEventsState(ctx, participantID)
		if err != nil {
			fmt.Println("Error occurred getting user events state:", err)
		}
		participant := BBBEvents.User{InternalUserID: dbUser.InternalUserID, ExternalUserID: dbUser.ExternalUserID, Name: dbUser.GdprName, Role: dbUser.Role, Guest: dbUser.IsGuest.Bool}

		for _, state := range states {
			switch state.EventType {
			case "user-left":
				if ts, ok := state.LatestTimestamp.(time.Time); ok {
					if participant.LeaveTimestamp == nil && !ts.IsZero() {
						participant.LeaveTimestamp = &ts
					}
				}
			case "user-presenter-as-signed":
				participant.Presenter = true
			case "user-audio-voice-enabled":
				participant.ListeningOnly = true
			case "user-audio-voice-disabled":
				participant.ListeningOnly = false
			case "user-audio-muted":
				participant.Muted = true
			case "user-audio-unmuted":
				participant.Muted = false
			case "user-cam-broadcast-start":
				participant.Stream = "broadcast"
			case "user-cam-broadcast-end":
				participant.Stream = ""
			case "user-emoji-changed":
			}
		}
		if !meeting.MeetingEnded.Time.IsZero() {
			participant.LeaveTimestamp = &meeting.MeetingEnded.Time
		}
		if ctx.Value(Gdpr("gdpr")).(bool) {
			participant.Name = dbUser.GdprName
		} else {
			participant.Name = dbUser.Name
		}
		participants = append(participants, participant)
	}

	return participants, err
}

func FillMeetingDetails(details []Detail, meeting db.Meeting) []Detail {
	// Load Details

	return details
}
