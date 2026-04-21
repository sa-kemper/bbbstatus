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
	"bbbstatus/locales"
	"bbbstatus/web"
	"fmt"
	"html/template"
	"net"
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/labstack/echo/v4"
	"github.com/labstack/echo/v4/middleware"
	"github.com/labstack/gommon/log"
)

var echof *echo.Echo

var defaultGohtmlTemplates = web.Views

var staticFS = web.StaticFS

func initEchoFramework(conf *config.ConfigurationStruct) {
	var Templates *Template
	const staticContentOverwriteFolder = "overwrite/static"
	const gohtmlTemplateOverwriteFolder = "overwrite/public/views"

	funcMap := template.FuncMap{
		"t": func(text string) string {
			return text
		},
		"reverse":        func(string, ...interface{}) string { return "x" },
		"formatTime":     func(t time.Time) string { return t.Format("15:04:05") },
		"formatDate":     func(t time.Time) string { return t.Format("2006.01.02 - 15:04") },
		"formatDuration": func(t time.Duration) string { return t.Round(time.Second).String() },
		"valFromIndex":   func(m map[string]int, s string) int { return m[s] },
		"timestamp":      func() string { return strconv.Itoa(int(time.Now().Unix())) },
		"timeDurationTsPtr": func(t time.Time, ptr *time.Time) string {
			return ptr.Sub(t).Round(time.Second).String()
		},
	}

	if _, err := os.Stat(gohtmlTemplateOverwriteFolder); err == nil {
		// Use files from the overwrite folder if it exists
		_, err := os.ReadDir(gohtmlTemplateOverwriteFolder)
		if err != nil {
			_, _ = fmt.Fprintf(os.Stderr, "Unable to read overwrite folder: %v\n", err)
			Templates = &Template{
				templates: template.Must(template.New("").Funcs(funcMap).ParseFS(defaultGohtmlTemplates, "views/*.gohtml")),
			}
		} else {
			// Template Setup with translation function
			Templates = &Template{
				templates: template.Must(template.New("").Funcs(funcMap).ParseGlob(gohtmlTemplateOverwriteFolder + "/*.gohtml")),
			}
		}
	} else {
		Templates = &Template{
			templates: template.Must(template.New("").Funcs(funcMap).ParseFS(defaultGohtmlTemplates, "views/*.gohtml")),
		}
	}

	// Echo Framework Routing setup
	//goland:noinspection SpellCheckingInspection
	echof = echo.New()
	echof.Renderer = Templates
	echof.Use(middleware.Logger())
	echof.Use(middleware.Recover())
	echof.Use(locales.AddTranslatorToContext)
	echof.Logger.SetLevel(log.INFO)

	// setup trusted proxies in order to trust x-forwarded-for.
	if len(conf.BaseConfig.TrustedProxies) > 3 {
		var proxyTrustOptions []echo.TrustOption
		for _, strCIDR := range strings.Split(conf.BaseConfig.TrustedProxies, ",") {
			if len(strCIDR) < 3 {
				panic("Something is seriously wrong with the TRUSTED_PROXIES configuration. CIDR read:" + strCIDR + ", Full config: " + conf.BaseConfig.TrustedProxies)
			}
			_, cidr, err := net.ParseCIDR(strCIDR)
			if err != nil {
				fmt.Println("Echo framework init -> proxy trust options")
				fmt.Printf("Error parsing CIDR: %v\n", err)
				panic(err)
			}
			proxyTrustOptions = append(proxyTrustOptions, echo.TrustIPRange(cidr))
		}
		echof.IPExtractor = echo.ExtractIPFromXFFHeader(proxyTrustOptions...)
	}

	if _, err := os.Stat(staticContentOverwriteFolder); err == nil {
		echof.Static("/static", staticContentOverwriteFolder)
	} else {
		fs := echo.MustSubFS(staticFS, "static/")
		echof.StaticFS("/static", fs)
	}
}
