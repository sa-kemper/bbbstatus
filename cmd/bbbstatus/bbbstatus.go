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
	"fmt"
	"html/template"
	"log"
	"math/rand"
	"time"

	"github.com/labstack/echo/v4/middleware"
	"github.com/nicksnyder/go-i18n/v2/i18n"
)

var FrontendTextMessages []i18n.Message
var err error

type Template struct {
	templates *template.Template
}

func main() {
	err = initDatabase()
	if err != nil {
		fmt.Println("[!] Error connecting to database, this is required for the operation of bbbstatus [!]")
		log.Fatalln(err)
		return
	}

	initI18n()          // initialize the localisation
	initEchoFramework() // initialize the "framework"
	initRoutes()        // start routing

	// regenerate randomness pool on startup.
	rand.NewSource(time.Now().Unix())

	echof.Logger.Fatal(echof.Start(confGet("HOST") + ":" + confGet("PORT")))
	echof.Pre(middleware.HTTPSNonWWWRedirect())
}
