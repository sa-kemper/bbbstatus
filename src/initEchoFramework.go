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
	"embed"
	"fmt"
	"github.com/labstack/echo/v4"
	"github.com/labstack/echo/v4/middleware"
	"github.com/labstack/gommon/log"
	"html/template"
	"net"
	"os"
	"strings"
)

var echof *echo.Echo

//go:embed public/views/*.gohtml
var defaultGohtmlTemplates embed.FS

//go:embed static
var staticFS embed.FS

func initEchoFramework() {
	var Templates *Template
	const staticContentOverwriteFolder = "overwrite/static"
	const gohtmlTemplateOverwriteFolder = "overwrite/public/views"

	funcMap := template.FuncMap{
		"t": func(text string) string {
			return text
		}, "reverse": func(string, ...interface{}) string { return "x" },
	}

	if _, err := os.Stat(gohtmlTemplateOverwriteFolder); err == nil {
		// Use files from the overwrite folder if it exists
		_, err := os.ReadDir(gohtmlTemplateOverwriteFolder)
		if err != nil {
			_, _ = fmt.Fprintf(os.Stderr, "Unable to read overwrite folder: %v\n", err)
			Templates = &Template{
				templates: template.Must(template.New("").Funcs(funcMap).ParseFS(defaultGohtmlTemplates, "public/views/*.gohtml")),
			}
		} else {
			// Template Setup with translation function
			Templates = &Template{
				templates: template.Must(template.New("").Funcs(funcMap).ParseGlob(gohtmlTemplateOverwriteFolder + "/*.gohtml")),
			}
		}
	} else {
		Templates = &Template{
			templates: template.Must(template.New("").Funcs(funcMap).ParseFS(defaultGohtmlTemplates, "public/views/*.gohtml")),
		}
	}

	// Echo Framework Routing setup
	//goland:noinspection SpellCheckingInspection
	echof = echo.New()
	echof.Renderer = Templates
	echof.Use(middleware.Logger())
	echof.Use(middleware.Recover())
	echof.Logger.SetLevel(log.INFO)

	// setup trusted proxies in order to trust x-forwarded-for.
	var proxyTrustOptions []echo.TrustOption
	for _, strCIDR := range strings.Split(confGet("TRUSTED_PROXIES"), ",") {
		_, cidr, err := net.ParseCIDR(strCIDR)
		if err != nil {
			fmt.Println("Echo framework init -> proxy trust options")
			fmt.Printf("Error parsing CIDR: %v\n", err)
			panic(err)
		}
		proxyTrustOptions = append(proxyTrustOptions, echo.TrustIPRange(cidr))
	}

	echof.IPExtractor = echo.ExtractIPFromXFFHeader(proxyTrustOptions...)

	if _, err := os.Stat(staticContentOverwriteFolder); err == nil {
		echof.Static("/static", staticContentOverwriteFolder)
	} else {
		fs := echo.MustSubFS(staticFS, "static/")
		echof.StaticFS("/static", fs)
	}
}
