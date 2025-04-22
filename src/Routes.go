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
	"github.com/labstack/echo/v4"
	"github.com/nicksnyder/go-i18n/v2/i18n"
	"net/http"
)

func init() { // Add all messages that are related to this file into the localization bundle
	var msgs = []i18n.Message{
		{ID: "404PageNotFoundHeader", Other: "404 - Page Not Found"},
		{ID: "PageNotFoundErrorMessage", Other: "Oops! The page you were looking for cannot be found."},
		{ID: "ReturnToIndexLinkText", Other: "Go to Homepage"},
		{ID: "404PageNotFoundHeader", Other: "404 - Page Not Found"},
	}
	FrontendTextMessages = append(FrontendTextMessages, msgs...)
}

type route struct {
	Name        string
	Method      string
	Path        string
	HandlerFunc func(echo.Context) error
}

var routes = []route{
	{Name: "meetings", Method: "GET", Path: "/meetings/", HandlerFunc: showMeetings},
	{Name: "report", Method: "GET", Path: "/report/:id", HandlerFunc: showMeetingReport},
	{Name: "reportCsv", Method: "GET", Path: "/report/:id/csv", HandlerFunc: downloadMeetingReport},
	{Name: "inspectReport", Method: "GET", Path: "/report/:id/inspect", HandlerFunc: showFilteredMeetingReport},
	{Name: "webhookEvent", Method: "POST", Path: "/event", HandlerFunc: bbbWebHookEvent},
	{Name: "statistics", Method: "GET", Path: "/statistics", HandlerFunc: statsPage},
	{Name: "statisticsCsv", Method: "GET", Path: "/statistics/csv", HandlerFunc: statsPageCSV},
	{Name: "index", Method: "GET", Path: "/", HandlerFunc: func(context echo.Context) error {
		return context.Redirect(http.StatusMovedPermanently, context.Echo().Reverse("meetings"))
	}},
	{Name: "catchAll", Method: "GET", Path: "/*", HandlerFunc: func(context echo.Context) error { return context.Render(http.StatusOK, "notfound", nil) }},
}

func initRoutes() {
	for _, route := range routes {
		if route.Method == "GET" {
			echof.GET(route.Path, route.HandlerFunc).Name = route.Name
		}
		if route.Method == "POST" {
			echof.POST(route.Path, route.HandlerFunc).Name = route.Name
		}
	}
}
