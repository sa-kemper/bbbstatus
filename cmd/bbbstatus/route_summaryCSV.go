package main

import (
	"bbbstatus/locales"
	"fmt"
	"net/http"
	"sort"
	"strings"
	"time"

	"github.com/labstack/echo/v4"
)

func summaryCSVHandler(c echo.Context) error {
	requestParams, Months, err := getSummaryOfMeetingsForDates(c)
	if err != nil {
		fmt.Println("error getting summary of meetings:" + err.Error())
		return err
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

	c.Response().Header().Set("Content-Disposition", "attachment; filename="+fmt.Sprintf("bbbstatus-%s-summary-%s.csv", fmt.Sprintf("StartDate(%v) - EndDate(%v)", requestParams.StartTime, requestParams.StopTime, time.Now().Format("2006-01-02"))))
	return c.Blob(http.StatusOK, "text/csv", []byte(csv))
}
