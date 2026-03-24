package main

import (
	db "bbbstatus/internal/database"
	"bbbstatus/locales"
	"context"
	"fmt"
	"net/http"
	"sort"
	"strings"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgtype"
	"github.com/labstack/echo/v4"
)

func summaryCSVHandler(c echo.Context) (err error) {
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
	var ctx = context.WithValue(c.Request().Context(), locales.Translator("Translator"), c.Get("Translator"))
	var conn *pgx.Conn

	conn, err = pgx.Connect(ctx, confGet("DB_CONNECTION_STRING"))
	if err != nil {
		fmt.Printf("ERROR Unable to connect to database: %v\n", err)
		return err
	}
	var dbQueries = db.New(conn)

	firstMeeting, err := dbQueries.GetFirstMeetingDate(ctx)
	if err != nil {
		fmt.Println("ERROR occurred whilst fetching first meeting date", err.Error())
	}
	lastMeeting, err := dbQueries.GetLastMeetingDate(ctx)
	if err != nil {
		fmt.Println("ERROR occurred whilst fetching last meeting date", err.Error())
	}
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
	var csv = strings.Join([]string{locales.TranslateFromEchoContext(c, "Month"), locales.TranslateFromEchoContext(c, "Time"), locales.TranslateFromEchoContext(c, "SummaryMeetingName"), locales.TranslateFromEchoContext(c, "SummaryUserCount"), locales.TranslateFromEchoContext(c, "SummaryDuration")}, ",") + "\n"
	// Sort the data for output
	sort.Slice(Months, func(i, j int) bool {
		iTime, _ := time.Parse("January", Months[i].Name)
		jTime, _ := time.Parse("January", Months[j].Name)
		return iTime.Before(jTime)
	})

	for iterator, month := range Months {
		sort.Slice(Months[iterator].Items, func(i, j int) bool {
			return Months[iterator].Items[i].Date.Before(Months[iterator].Items[j].Date)
		})
		csv += fmt.Sprintf("%s,%s,%s,%s,%s\n", month.Name, "", "", "", "")
		for _, monthItem := range month.Items {
			csv += fmt.Sprintf("%s,%s,%s,%d,%s\n", monthItem.Date.Format("2006.01.02"), monthItem.Date.Format("15:04:05"), monthItem.MeetingName, monthItem.UserCount, monthItem.Duration.String())
		}
	}

	c.Response().Header().Set("Content-Disposition", "attachment; filename="+fmt.Sprintf("bbbstatus-%s-summary-%s.csv", fmt.Sprintf("StartDate(%v) - EndDate(%v)", userInputStartTime.Format("2006.01.02"), userInputStopTime.Format("2006.01.02")), time.Now().Format("2006-01-02")))
	return c.Blob(http.StatusOK, "text/csv", []byte(csv))
}
