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
	"github.com/jackc/pgx/v5"
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
	TrustedProxies     string `toml:"TRUSTED_PROXIES"`
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

func (s bbbServer) isEqual(other bbbServer) bool {
	return s.Hostname == other.Hostname &&
		s.ApiPort == other.ApiPort &&
		s.SharedSecret == other.SharedSecret &&
		s.FriendlyName == other.FriendlyName
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

var adminConf config
var runtimeBbbServers []bbbServer // runtimeBbbServers is used to change configured bbbServer's at runtime (api key change, user settings, etc.) without mutating the original adminConf.

func init() {
	var configFileConf config
	var content, err = os.ReadFile("config.toml")

	if err != nil { // error reading the config
		// TODO: handle the error differently depending on why the file couldn't be read.
		fmt.Println("Error reading config.toml:", err) // the config file may be unreadable. further error handling is required.
		defaultConfig, err := toml.Marshal(defaultConfiguration)
		if err != nil {
			fmt.Println("Error occurred marshaling the default config file to the current work directory:", err) // unlikely during production, this is mainly for development.
			return
		}
		fmt.Println("Attempting to write default config file to the current work directory")
		err = os.WriteFile("config.toml", defaultConfig, 0600)
		if err != nil {
			fmt.Println("Error occurred writing the default config file to the current work directory:", err)
		}
	}

	if content != nil {
		err = toml.Unmarshal(content, &configFileConf)
		if err != nil {
			fmt.Println("Error occurred unmarshalling the default config file to the current work directory:", err)
		}
		adminConf = configFileConf
	}

	parseBBBServersFromEnv()

	if len(adminConf.BBBServers) < 1 && len(runtimeBbbServers) < 1 {
		panic("No bbb servers configured.")
	} else {
		ValidateConfiguredBBBServers()
	}
}

func parseBBBServersFromEnv() {
	bbbServersEnv, ok := os.LookupEnv("BBB_SERVERS")
	if !ok && len(adminConf.BBBServers) < 1 { // no bbbServers were configured.
		panic("BBB_SERVERS environment must be set in order for bbbstatus to work.")
	}
	if !ok { // config file was used
		fmt.Println("BBB_SERVERS environment variable not set", bbbServersEnv, ok)
		return
	}

	// remove potential artifacts from the adminConf.
	bbbServersEnv = strings.ReplaceAll(bbbServersEnv, "\"", "")
	bbbServersEnv = strings.ReplaceAll(bbbServersEnv, "'", "")
	bbbServersEnv = strings.TrimSpace(bbbServersEnv)
	bbbServersEnv = strings.ToLower(bbbServersEnv)
	conn, err := pgx.Connect(context.TODO(), confGet("DB_CONNECTION_STRING"))
	if err != nil {
		fmt.Println("error parseBBBServersFromEnv -> connecting to database:", err)
		return
	}
	defer conn.Close(context.TODO())

	for _, serverUrl := range strings.Split(bbbServersEnv, ",") { // serverUrl is expected to be hostname:port, omitting the protocol as it's https in most cases. port and colon are optional
		var serverHostname, serverPort string
		if strings.Contains(serverUrl, ":") {
			urlParts := strings.Split(serverUrl, ":")
			if len(urlParts) > 2 {
				panic("Malformed BBB_SERVERS URL")
			}
			serverHostname = urlParts[0]
			serverPort = urlParts[1]
		} else {
			serverHostname = serverUrl
			serverPort = "443"
		}
		//fmt.Println("DEBUG: parseBBBServersFromEnv->serverUrl:'" + serverUrl + "', ServerHostname:'" + serverHostname + "', ServerPort:'" + serverPort + "'")
		err = addBBBServerToDb(conn, bbbServer{Hostname: serverHostname, ApiPort: serverPort})
		if err != nil {
			fmt.Println("Error occurred adding BBB server to database:", err)
		}
	}
}

func ValidateConfiguredBBBServers() {
	// Validate the read BBBServers config by checking for API access
	for _, server := range adminConf.BBBServers {
		if server.SharedSecret == "nil" { // database default for nil value.
			continue
		}
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
		if val, exists := os.LookupEnv(key); exists {
			return val
		}
		if adminConf.BaseConfig.Host != "" {
			return adminConf.BaseConfig.Host
		}
		if host, _ = os.Hostname(); host != "" {
			return host
		}
		return "0.0.0.0"

	case "PORT":
		if val, exists := os.LookupEnv(key); exists {
			return val
		}
		if adminConf.BaseConfig.Port != "" {
			return adminConf.BaseConfig.Port
		}
		return "8080"

	case "DB_CONNECTION_STRING":
		if val, exists := os.LookupEnv(key); exists {
			return val
		}

		connStr := adminConf.DatabaseConfig.DatabaseConnectionString
		if connStr != "" {
			return connStr
		}

		fmt.Println("https://www.postgresql.org/docs/12/libpq-connect.html#id-1.7.3.8.3.6")
		fmt.Println("DB_CONNECTION_STRING example: postgres://bbbstatus:bbbstatus@localhost/bbbstatus")
		panic("DB_CONNECTION_STRING environment variable not set")

	case "CSV_STRUCTURE":
		if val, exists := os.LookupEnv(key); exists {
			return val
		}

		if csvStructure := adminConf.ReportConfig.CsvStructure; csvStructure != "" {
			valid := ValidateCSVStructureConfig(csvStructure)
			if valid != nil {
				panic("The provided CSV_STRUCTURE config.toml variable is not valid.")
			}
		}

		return "time,user,action,text representation"

	case "SERVE_STATIC_CONTENT":
		if val, exists := os.LookupEnv(key); exists {
			return strings.ToLower(val)
		}

		if adminConf.BaseConfig.ServeStaticContent {
			return "true"
		}
		return "false"
	case "TRUSTED_PROXIES":
		if val, exists := os.LookupEnv(key); exists {
			return val
		}
		if trustedProxies := adminConf.BaseConfig.TrustedProxies; trustedProxies != "" {
			return trustedProxies
		}
		return "127.0.0.1/8,172.17.0.1/16,::1/128" // In my opinion, a sane default.
	}

	fmt.Printf("ISSUES REGARDING CONFIGURATION KEY: '%s'\r\n", key)
	return os.Getenv(key)
}

// todo: run go routine that reloads the adminConf content on a OS signal

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
