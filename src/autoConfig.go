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
	"github.com/BurntSushi/toml"
	"os"
	"slices"
	"strings"
)

var conf = map[string]string{
	"PORT":                 "8080",
	"HOST":                 "localhost",
	"DB_CONNECTION_STRING": "postgres://myapp:coolAppPassword!1@localhost/appdb",
	"CSV_STRUCTURE":        "time,user,action,text representation",
	"PROTOCOL":             "http",
	"SSL_KEY":              "",
	"SSL_CERT":             "",
	"SERVE_STATIC_CONTENT": "true",
}
var confFileExists bool

func init() {
	var configFileConf = make(map[string]string)
	var content, err = os.ReadFile("config.toml")
	if err != nil {
		fmt.Println("config.toml does not exist")
		confFileExists = false
		defaultConfig, err := toml.Marshal(conf)
		if err != nil {
			fmt.Println("Error occurred marshaling the default config file to the current work directory:", err)
			return
		}
		fmt.Println("Attempting to write default config file to the current work directory")
		err = os.WriteFile("config.toml", defaultConfig, 0600)
		if err != nil {
			fmt.Println("Error occurred writing the default config file to the current work directory:", err)
			return
		}
		return
	}
	err = toml.Unmarshal(content, &configFileConf)
	if err != nil {
		fmt.Println("Error occurred unmarshaling the default config file to the current work directory:", err)
	}
	confFileExists = true
	conf = configFileConf
}

// A function to safely obtain configured values and sensible defaults
func confGet(key string) string {
	if confFileExists {
		return conf[key]
	}
	switch key {
	case "HOST":
		var host string
		if host = os.Getenv("HOST"); host != "" {
			return host
		}
		if host, _ = os.Hostname(); host != "" {
			return host
		}
		return "localhost"

	case "PORT":
		if port := os.Getenv("PORT"); port != "" {
			return port
		}
		return "8080"

	case "DB_CONNECTION_STRING":
		connStr := os.Getenv("DB_CONNECTION_STRING")
		if connStr == "" {
			fmt.Println("https://www.postgresql.org/docs/12/libpq-connect.html#id-1.7.3.8.3.6")
			fmt.Println("DB_CONNECTION_STRING example: postgres://myapp:coolAppPassword!1@localhost/appdb")
			panic("DB_CONNECTION_STRING environment variable not set")
		}

	case "CSV_STRUCTURE":
		if csvStructure := os.Getenv("CSV_STRUCTURE"); csvStructure != "" {
			valid := ValidateCSVStructureConfig(csvStructure)
			if valid != nil {
				panic(valid)
			}
			return csvStructure
		}
		return "time,user,action,text representation"

	case "SERVE_STATIC_CONTENT":
		var serveStaticContent = os.Getenv("SERVE_STATIC_CONTENT")
		if strings.ToLower(strings.TrimSpace(serveStaticContent)) == "true" {
			return "true"
		}
		if strings.ToLower(strings.TrimSpace(serveStaticContent)) == "false" {
			return "false"
		}

		fmt.Printf("we have trouble understanding your configuration. Please check the content of the SERVE_STATIC_CONTENT environment variable.\n")
		return "true"
	}
	fmt.Printf("ISSUES REGARDING CONFIGURATION KEY: '%s'\r\n", key)
	return os.Getenv(key)
}

// todo: run go routine that reloads the conf content on a OS signal

func ValidateCSVStructureConfig(csvStructure string) error {
	// Validate CSV Struct config
	var csvStructureConfig = strings.Split(csvStructure, ",")
	var validCsvStructureObjects = []string{"time", "user", "action", "textrepresentation"}
	for _, csvObject := range csvStructureConfig {
		csvObject = strings.TrimSpace(csvObject)
		csvObject = strings.ReplaceAll(csvObject, " ", "")
		csvObject = strings.ToLower(csvObject)

		var valid = slices.Contains(validCsvStructureObjects, csvObject)
		if !valid {
			fmt.Println("Error parsing the CSV_STRUCTURE configuration")
			return fmt.Errorf("invalid CSVStructureObject: %s", csvObject)
		}
	}
	return nil
}
