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
	db "bbbstatus/internal/database"
	"errors"
	"fmt"
	"github.com/jackc/pgx/v5"
	"github.com/labstack/echo/v4"
	"net/http"
	"slices"
	"time"
)

func showFilteredMeetingReport(context echo.Context) error {
	type customUserType struct {
		InternalUserID string
		ExternalUserID string
		Name           string
		Role           string
		Presenter      bool
		Guest          bool
		ListeningOnly  bool
		SharingMic     bool
		Muted          bool
		Stream         string
		RaiseHand      bool
		Emoji          string
		LeaveTimestamp *time.Time
		IsFilteredFor  bool
	}
	var internalMeetingId = context.Param("id")
	var customParticipants []customUserType
	var filteredForUserIds []string

	// use the context provided by the user request in order to respect the browsers / servers / admins settings.
	var ctx = context.Request().Context()
	conn, err := pgx.Connect(ctx, confGet("DB_CONNECTION_STRING"))
	if err != nil {
		_ = context.Render(http.StatusInternalServerError, "errorPage", map[string]interface{}{"ErrorTitle": "Internal Error", "ErrorParagraph": err.Error()})
		return err
	}
	defer conn.Close(ctx)

	dbQueries := db.New(conn)

	// get the participants upfront so we can filter them out in the GenerateWebReport function
	// this ensures that only predictable url parameters are ever parsed, the user could provide user id's that do not exist, which might cause issues somewhere sometimes.
	participants, err := dbQueries.GetUsersInMeetingByInternalID(ctx, internalMeetingId)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return context.Render(http.StatusNotFound, "notfound", nil)
		}
		_ = context.Render(http.StatusInternalServerError, "errorPage", map[string]interface{}{"ErrorTitle": "Internal Error", "ErrorParagraph": err.Error()})
		return err
	}

	for _, participant := range participants {
		if context.QueryParam(participant) == "on" {
			filteredForUserIds = append(filteredForUserIds, participant)
		}
	}

	report, err := GenerateWebReport(ctx, internalMeetingId, &filteredForUserIds)
	if err != nil {
		fmt.Println("error occurred when generating report", err)
		return err
	}

	// we want to convert the types so we can add additional properties for the template (IsFilteredFor)
	for _, participant := range report.Participants {
		customParticipants = append(customParticipants, customUserType{
			InternalUserID: participant.InternalUserID,
			ExternalUserID: participant.ExternalUserID,
			Name:           participant.Name,
			Role:           participant.Role,
			Presenter:      participant.Presenter,
			Guest:          participant.Guest,
			ListeningOnly:  participant.ListeningOnly,
			SharingMic:     participant.SharingMic,
			Muted:          participant.Muted,
			Stream:         participant.Stream,
			RaiseHand:      participant.RaiseHand,
			Emoji:          participant.Emoji,
			LeaveTimestamp: participant.LeaveTimestamp,
			IsFilteredFor:  slices.Contains(filteredForUserIds, participant.InternalUserID),
		})
	}

	filteredReport := struct {
		Details      []Detail
		Participants []customUserType
		Timeline     []Event
		Recordings   []BBBAPI.Recording
	}{Details: report.Details, Participants: customParticipants, Timeline: report.Timeline, Recordings: report.Recordings}
	return context.Render(http.StatusOK, "inspectReport", map[string]interface{}{"InternalMeetingID": internalMeetingId, "Report": filteredReport})
}
