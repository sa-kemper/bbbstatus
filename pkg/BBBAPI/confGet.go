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
	"os"
	"strings"
)

func confGet(name string) string {
	var env = make(map[string]string)
	envConf, err := os.ReadFile(".env")
	if err != nil {
		panic(err)
	}

	for _, v := range strings.Split(string(envConf), "\n") {
		keyVal := strings.Split(v, "=")
		if len(keyVal) != 2 {
			continue
		}
		env[keyVal[0]] = keyVal[1]
	}

	if env[name] == "" {
		return os.Getenv(name)
	}
	return env[name]
}
