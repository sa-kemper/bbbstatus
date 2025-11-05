package main

import (
	db "bbbstatus/internal/database"
	"bbbstatus/pkg/BBBEvents"
	"context"
	"fmt"
	"net/http"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgtype"
	"github.com/labstack/echo/v4"
	"github.com/nicksnyder/go-i18n/v2/i18n"
)

func init() { // Add all messages that are related to this file into the localization bundle
	var msgs = []i18n.Message{
		{ID: "Summary", Other: "Summary"},
		{ID: "January", Other: "January"},
		{ID: "February", Other: "February"},
		{ID: "March", Other: "March"},
		{ID: "April", Other: "April"},
		{ID: "May", Other: "May"},
		{ID: "June", Other: "June"},
		{ID: "July", Other: "July"},
		{ID: "August", Other: "August"},
		{ID: "September", Other: "September"},
		{ID: "October", Other: "October"},
		{ID: "November", Other: "November"},
		{ID: "December", Other: "December"},
		{ID: "StatsPageSelectScopeStartTime", Other: "select start time"},
		{ID: "StatsPageSelectScopeStopTime", Other: "select stop time"},
		{ID: "StatsPageSelectScopeStartTimePrint", Other: "Start time"},
		{ID: "StatsPageSelectScopeStopTimePrint", Other: "Stop time"},
		{ID: "Time", Other: "Time"},
		{ID: "SummaryMeetingName", Other: "Meeting name"},
		{ID: "SummaryUserCount", Other: "User count"},
		{ID: "SummaryDuration", Other: "Duration"},
	}
	FrontendTextMessages = append(FrontendTextMessages, msgs...)
	for _, m := range BBBEvents.UserEventTextRepresentation { // Add user events text representation to the language strings.
		FrontendTextMessages = append(FrontendTextMessages, m)
	}
}

func summaryPage(c echo.Context) (err error) {

	type SummaryItem struct {
		Date        time.Time
		MeetingName string
		UserCount   int
		Duration    time.Duration
	}
	type Month struct {
		Name  string
		Items []*SummaryItem
	}
	var ctx = context.WithValue(c.Request().Context(), "Translator", c.Get("Translator"))
	var conn *pgx.Conn

	conn, err = pgx.Connect(ctx, confGet("DB_CONNECTION_STRING"))
	if err != nil {
		fmt.Printf("ERROR Unable to connect to database: %v\n", err)
		return err
	}
	var dbQueries = db.New(conn)

	firstMeeting, err := dbQueries.GetFirstMeetingDate(ctx)
	lastMeeting, err := dbQueries.GetLastMeetingDate(ctx)
	var requestParams = struct {
		StartTime    string `query:"startTime"`
		StopTime     string `query:"stopTime"`
		StartTimeMin string
		StartTimeMax string
	}{
		StartTime:    c.QueryParam("startTime"),
		StopTime:     c.QueryParam("stopTime"),
		StartTimeMin: firstMeeting.(time.Time).Format("2006-01-02"),
		StartTimeMax: lastMeeting.(time.Time).Format("2006-01-02"),
	}

	// Parse user input, and correct its errors
	userInputStartTime, err := time.Parse("2006-01-02", requestParams.StartTime)
	if err != nil {
		userInputStartTime = firstMeeting.(time.Time) // default back to the first meeting on invalid input
	}
	userInputStopTime, err := time.Parse("2006-01-02", requestParams.StopTime)
	if err != nil {
		userInputStopTime = lastMeeting.(time.Time)
	}

	MeetingSlice, err := dbQueries.GetMeetingsBetweenDates(ctx, db.GetMeetingsBetweenDatesParams{
		CreateTime:   pgtype.Timestamp{Time: userInputStopTime, Valid: true},
		CreateTime_2: pgtype.Timestamp{Time: userInputStartTime, Valid: true},
	})
	if err != nil {
		fmt.Println("error occurred whilst fetching meeting between dates", err.Error())
		return err
	}

	var Months []Month
	var MonthMap = make(map[string]*Month)
	for _, meeting := range MeetingSlice {
		meetingCreateMonthString := meeting.CreateTime.Time.Month().String()
		monthPointer, ok := MonthMap[meetingCreateMonthString]
		usersFromMeeting, err := dbQueries.GetUserCountInMeetingByInternalID(ctx, meeting.InternalMeetingID)
		if err != nil {
			fmt.Println("error occurred whilst fetching user count in meeting", err.Error())
		}
		if !ok {
			MonthMap[meetingCreateMonthString] = &Month{
				Name: meetingCreateMonthString,
				Items: []*SummaryItem{{
					Date:        meeting.CreateTime.Time,
					MeetingName: meeting.Name,
					UserCount:   int(usersFromMeeting),
					Duration:    meeting.MeetingEnded.Time.Sub(meeting.CreateTime.Time),
				}},
			}
			continue
		}
		monthPointer.Items = append(monthPointer.Items, &SummaryItem{
			Date:        meeting.CreateTime.Time,
			MeetingName: meeting.Name,
			UserCount:   int(usersFromMeeting),
			Duration:    meeting.MeetingEnded.Time.Sub(meeting.CreateTime.Time),
		})
	}
	for _, month := range MonthMap {
		Months = append(Months, *month)
	}

	return c.Render(http.StatusOK, "summary", map[string]interface{}{"Data": Months, "Request": requestParams})
}
