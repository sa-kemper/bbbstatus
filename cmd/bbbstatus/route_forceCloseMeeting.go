package main

import (
	"bbbstatus/internal/database"
	"fmt"
	"net/http"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgtype"
	"github.com/labstack/echo/v4"
)

func forceCloseMeeting(context echo.Context) error {
	var ctx = context.Request().Context()
	// Connect to the db using
	conn, err := pgx.Connect(ctx, confGet("DB_CONNECTION_STRING"))
	if err != nil {
		return fmt.Errorf("unable to connect to database: %v", err)
	}
	//goland:noinspection ALL
	defer conn.Close(ctx)

	var dbtx = bbbstatus.New(conn)
	err = dbtx.EndMeetingAtTimestampByID(ctx, bbbstatus.EndMeetingAtTimestampByIDParams{MeetingEnded: pgtype.Timestamp{Time: time.Now(), Valid: true}, InternalMeetingID: context.Param("id")})
	if err != nil {
		fmt.Println("error occured during forceCloseMeeting:", err)
		return err
	}
	return context.Redirect(http.StatusOK, context.Echo().Reverse("report", context.Param("id")))
}
