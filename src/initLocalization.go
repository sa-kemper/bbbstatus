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
	"bbbstatus/internal/BBBEvents"
	"embed"
	"fmt"
	"github.com/BurntSushi/toml"
	"github.com/nicksnyder/go-i18n/v2/i18n"
	"golang.org/x/text/language"
	"os"
)

//go:embed locales/active.*.toml
var defaultLocaleFile embed.FS
var Bundle *i18n.Bundle
var localizer *i18n.Localizer

func init() { // Add all messages that are related to this file into the localization bundle
	var msgs = []i18n.Message{
		{ID: "CurrentISO639", Other: "en"},
	}
	FrontendTextMessages = append(FrontendTextMessages, msgs...)
	for _, m := range BBBEvents.UserEventTextRepresentation { // Add user events text representation to the language strings.
		FrontendTextMessages = append(FrontendTextMessages, m)
	}
}

func initI18n() {
	Bundle = i18n.NewBundle(language.English)

	for _, m := range FrontendTextMessages {
		err = Bundle.AddMessages(language.English, &m)
		if err != nil {
			_, _ = fmt.Fprintf(os.Stderr, "Unable to add message: %v\n", err)
		}
	}

	Bundle.RegisterUnmarshalFunc("toml", toml.Unmarshal)
	overwriteFolder := "overwrite/locales"

	if _, err := os.Stat(overwriteFolder); err == nil {
		// Use files from the overwrite folder if it exists
		files, err := os.ReadDir(overwriteFolder)
		if err != nil {
			_, _ = fmt.Fprintf(os.Stderr, "Unable to read overwrite folder: %v\n", err)
		} else {
			for _, file := range files {
				if !file.IsDir() {
					path := overwriteFolder + "/" + file.Name()
					_, err = Bundle.LoadMessageFile(path)
					if err != nil {
						_, _ = fmt.Fprintf(os.Stderr, "Unable to load message file: %v\n", err)
					}
				}
			}
		}
	} else {
		// Walk the embedded locales and load each of them.
		files, err := defaultLocaleFile.ReadDir("locales")
		if err != nil {
			_, _ = fmt.Fprintf(os.Stderr, "Unable to read locales: %v\n", err)
		}
		for _, file := range files {
			if file.IsDir() {
				continue
			}
			_, err = Bundle.LoadMessageFileFS(defaultLocaleFile, "locales/"+file.Name())
			if err != nil {
				_, _ = fmt.Fprintf(os.Stderr, "Unable to load embedded message file: %v\n", err)
			}
		}
	}
}
