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
	"github.com/labstack/echo/v4/middleware"
	"github.com/nicksnyder/go-i18n/v2/i18n"
	"golang.org/x/text/language"
	"html/template"
	"io"
	"net"
	"regexp"
	"strings"
)

var FrontendTextMessages []i18n.Message
var err error

type Template struct {
	templates *template.Template
}

// Render is a function hook to provide function calls inside html templates, such as Translate
func (t *Template) Render(w io.Writer, name string, data interface{}, c echo.Context) error {
	// Genius move right there, not because it's hard but because I've not seen this documented at the time of writing.
	t.templates.Funcs(template.FuncMap{
		"t": func(text string) string {
			localizer = i18n.NewLocalizer(Bundle, c.Request().Header.Get("Accept-Language"), language.English.String())
			return Translate(text)

		},
		"reverse": c.Echo().Reverse,
	})
	return t.templates.ExecuteTemplate(w, name, data)
}

// Translate is a simple i18n function that translates text and fails at runtime if the string is not available in the requested language.
// failing hard makes testing the code a lot easier.
func Translate(text string) string {
	re := regexp.MustCompile("CW(?i)[0-9]{1,2}")
	matched := re.MatchString(text)
	if matched {
		splitText := strings.Split(text, "CW")
		if len(splitText) > 1 {
			return localizer.MustLocalize(&i18n.LocalizeConfig{MessageID: "CW"}) + splitText[1]
		}
	}
	return localizer.MustLocalize(&i18n.LocalizeConfig{MessageID: text})
}

// TranslateAdvanced is a simple i18n function that supports text templating
func TranslateAdvanced(text string, data map[string]string) string {
	return localizer.MustLocalize(&i18n.LocalizeConfig{MessageID: text, TemplateData: data})
}

func getIpFromContext(c echo.Context) net.IP {
	contextIp := c.RealIP()
	if strings.Contains(contextIp, "[") {
		endIndex := strings.Index(contextIp, "]:")
		return net.ParseIP(contextIp[1:endIndex])
	}
	if strings.Contains(contextIp, ":") {
		endIndex := strings.Index(contextIp, ":")
		return net.ParseIP(contextIp[:endIndex])
	}
	return net.ParseIP(contextIp)
}

func main() {
	err = initDatabase()
	if err != nil {
		panic(err)
	}
	initI18n()
	initEchoFramework()
	initRoutes()

	echof.Logger.Fatal(echof.Start(confGet("HOST") + ":" + confGet("PORT")))
	echof.Pre(middleware.HTTPSNonWWWRedirect())
}
