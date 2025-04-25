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

package namegen

import (
	"embed"
	"errors"
	"fmt"
	"io/fs"
	"log"
	"math/rand"
	"slices"
	"strings"
)

//go:embed assets
var assets embed.FS

func Generate(lang string) (name string) {
	adjectives, err := assets.ReadFile(fmt.Sprintf("assets/adjective_%s.txt", lang))
	if err != nil {
		if errors.Is(err, fs.ErrNotExist) {
			log.Fatal("error occurred while reading adjectives file, language " + lang + " does not exist.")
		}
		log.Fatal(err)
	}

	animals, err := assets.ReadFile(fmt.Sprintf("assets/animals_%s.txt", lang))
	if err != nil {
		if errors.Is(err, fs.ErrNotExist) {
			log.Fatal("error occurred while reading animals file, language " + lang + " does not exist.")
		}
		log.Fatal(err)
	}
	adjArray := strings.Split(string(adjectives), "\n")
	aniArray := strings.Split(string(animals), "\n")

	name += adjArray[rand.Intn(len(aniArray)-1)]
	name += aniArray[rand.Intn(len(aniArray)-1)]

	return name
}

func GenerateUnique(lang string, existingNames []string) (name string) {
	var attemptCounter int

	for {
		var generatedName string
		if attemptCounter < 50 {
			generatedName = Generate(lang)
		} else {
			charset := "abcdefghijklmnopqrstuvwxyz"
			for i := 0; i < 6; i++ {
				generatedName += string(charset[rand.Intn(len(charset)-1)])
			}
		}

		if attemptCounter > 500 {
			log.Fatal("failed to generate unique name")
		}
		if !slices.Contains(existingNames, generatedName) {
			return generatedName
		}
	}
}

func IsGeneratedName(lang, name string) bool {
	adjectives, err := assets.ReadFile(fmt.Sprintf("assets/adjective_%s.txt", lang))
	if err != nil {
		if errors.Is(err, fs.ErrNotExist) {
			log.Fatal("error occurred while reading adjectives file, language " + lang + " does not exist.")
		}
		log.Fatal(err)
	}

	animals, err := assets.ReadFile(fmt.Sprintf("assets/animals_%s.txt", lang))
	if err != nil {
		if errors.Is(err, fs.ErrNotExist) {
			log.Fatal("error occurred while reading animals file, language " + lang + " does not exist.")
		}
		log.Fatal(err)
	}
	adjArray := strings.Split(string(adjectives), "\n")
	aniArray := strings.Split(string(animals), "\n")

	for _, animal := range aniArray {
		if strings.Contains(strings.ToLower(name), strings.ToLower(animal)) {
			pos := strings.Index(strings.ToLower(name), strings.ToLower(animal))
			return slices.Contains(adjArray, name[0:pos])
		}
	}
	return false
}
