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

package BBBAPI

import (
	"crypto/sha1"
	"encoding/hex"
	"fmt"
	"sort"
	"strings"
)

func generateURL(config URLConfig) string {
	var paramString string
	if config.ApiPort != nil {
		// ensure port prefix.
		if !strings.HasPrefix(*config.ApiPort, ":") {
			*config.ApiPort = ":" + (*config.ApiPort)
		}
	} else {
		// prevent nil pointer dereference
		var empty string
		config.ApiPort = &empty
	}

	result := fmt.Sprintf("https://%s%s/bigbluebutton/api/%s?", config.Hostname, *config.ApiPort, config.Methode)

	keys := make([]string, 0, len(config.Parameters))
	for k := range config.Parameters {
		keys = append(keys, k)
	}
	sort.Strings(keys)

	for _, key := range keys {
		result += fmt.Sprintf("%s=%s&", key, config.Parameters[key])
		paramString += fmt.Sprintf("%s=%s&", key, config.Parameters[key])
	}

	checksumString := config.Methode + strings.TrimRight(paramString, "&") + config.SharedSecret
	checksum := sha1.New()
	checksum.Write([]byte(checksumString))
	return fmt.Sprintf("%schecksum=%s", result, hex.EncodeToString(checksum.Sum(nil)))
}
