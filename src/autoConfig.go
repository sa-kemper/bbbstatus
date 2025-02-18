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
	"BbbStatus/internal/BBBAPI"
	"context"
	"fmt"
	"github.com/BurntSushi/toml"
	"os"
	"slices"
	"strings"
	"time"
)

type config struct {
	BaseConfig     baseConfig
	ReportConfig   reportConfig
	DatabaseConfig dbConfig
	BBBServers     []bbbServer
}

type baseConfig struct {
	Host               string `toml:"HOST"`
	Port               string `toml:"PORT"`
	ServeStaticContent bool   `toml:"SERVE_STATIC_CONTENT"`
}

type reportConfig struct {
	CsvStructure string `toml:"CSV_STRUCTURE"`
}

type dbConfig struct {
	DatabaseConnectionString string `toml:"DB_CONNECTION_STRING"`
}

type bbbServer struct {
	Hostname     string `toml:"HOSTNAME"`      // required
	ApiPort      string `toml:"API_PORT"`      // required only if not 443
	SharedSecret string `toml:"SHARED_SECRET"` // required only if API usage is needed
	APITimeout   int    `toml:"API_TIMEOUT"`   // defaults to 2 (seconds)
	FriendlyName string `toml:"FRIENDLY_NAME"` // not required
}

var defaultConfiguration = config{
	BaseConfig:     baseConfig{Host: "0.0.0.0", Port: "8080", ServeStaticContent: true},
	ReportConfig:   reportConfig{CsvStructure: "time,user,action,text representation"},
	DatabaseConfig: dbConfig{DatabaseConnectionString: "postgres://bbbstatus:bbbstatus@localhost/bbbstatus"},
	BBBServers: []bbbServer{
		{Hostname: "localhost", ApiPort: "443", SharedSecret: "aDefaultSharedSecret", APITimeout: 5},
		{Hostname: "other-host", ApiPort: "443", SharedSecret: "aDefaultSharedSecret", APITimeout: 2},
	},
}

var conf config

func init() {
	var configFileConf config
	var content, err = os.ReadFile("config.toml")

	if err != nil { // error reading the config
		// TODO: handle the error differently depending on why the file couldn't be read.
		fmt.Println("config.toml does not exist") // the config file may be unreadable. further error handling is required.
		defaultConfig, err := toml.Marshal(defaultConfiguration)
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
		fmt.Println("Error occurred unmarshalling the default config file to the current work directory:", err)
	}

	conf = configFileConf

	if len(conf.BBBServers) < 1 {
		fmt.Println("No servers configured. Expect API config to not work at all, The webhook validator will consider any package as valid.")
	}
	// Validate the read BBBServers config by checking for API access
	for _, server := range conf.BBBServers {
		APITimeout := time.Duration(server.APITimeout) * time.Second
		var api BBBAPI.API
		if APITimeout != 0 {
			api = BBBAPI.API{Hostname: server.Hostname, Port: server.ApiPort, SharedSecret: server.SharedSecret, Timeout: &APITimeout}
		} else {
			api = BBBAPI.API{Hostname: server.Hostname, Port: server.ApiPort, SharedSecret: server.SharedSecret}
		}
		valid, err := api.ValidateApiSettings(context.Background())
		if !valid {
			fmt.Println(server.Hostname + ": Invalid API settings")
			fmt.Println("Error occurred validating the API settings: ", err)
		}
	}
}

// A function to safely obtain configured values and sensible defaults
func confGet(key string) string {
	switch key {
	case "HOST":
		var host string
		if host = os.Getenv("HOST"); host != "" {
			return host
		}
		if conf.BaseConfig.Host != "" {
			return conf.BaseConfig.Host
		}
		if host, _ = os.Hostname(); host != "" {
			return host
		}
		return "0.0.0.0"

	case "PORT":
		if port := os.Getenv("PORT"); port != "" {
			return port
		}
		if conf.BaseConfig.Port != "" {
			return conf.BaseConfig.Port
		}
		return "8080"

	case "DB_CONNECTION_STRING":
		connStr := os.Getenv("DB_CONNECTION_STRING")
		if connStr != "" {
			return connStr
		}

		connStr = conf.DatabaseConfig.DatabaseConnectionString
		if connStr != "" {
			return connStr
		}

		fmt.Println("https://www.postgresql.org/docs/12/libpq-connect.html#id-1.7.3.8.3.6")
		fmt.Println("DB_CONNECTION_STRING example: postgres://bbbstatus:bbbstatus@localhost/bbbstatus")
		panic("DB_CONNECTION_STRING environment variable not set")

	case "CSV_STRUCTURE":
		if csvStructure := os.Getenv("CSV_STRUCTURE"); csvStructure != "" {
			valid := ValidateCSVStructureConfig(csvStructure)
			if valid != nil {
				panic("The provided CSV_STRUCTURE environment variable is not valid.")
			}
			return csvStructure
		}
		if csvStructure := conf.ReportConfig.CsvStructure; csvStructure != "" {
			valid := ValidateCSVStructureConfig(csvStructure)
			if valid != nil {
				panic("The provided CSV_STRUCTURE config.toml variable is not valid.")
			}
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

		if conf.BaseConfig.ServeStaticContent {
			return "true"
		}
		return "false"
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
