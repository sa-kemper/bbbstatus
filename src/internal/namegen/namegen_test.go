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

import "testing"

func Test_isGeneratedName(t *testing.T) {
	type args struct {
		name string
		lang string
	}
	tests := []struct {
		name string
		args args
		want bool
	}{
		{name: "BaseCase1", args: args{name: Generate("de"), lang: "de"}, want: true},
		{name: "BaseCase2", args: args{name: Generate("de"), lang: "de"}, want: true},
		{name: "BaseCase3", args: args{name: Generate("de"), lang: "de"}, want: true},
		{name: "BaseCase4", args: args{name: Generate("de"), lang: "de"}, want: true},
		{name: "BaseCase5", args: args{name: Generate("de"), lang: "de"}, want: true},
		{name: "BaseCase6", args: args{name: Generate("de"), lang: "de"}, want: true},
		{name: "BaseCase7", args: args{name: Generate("de"), lang: "de"}, want: true},
		{name: "BaseCase8", args: args{name: Generate("de"), lang: "de"}, want: true},
		{name: "BaseCase9", args: args{name: Generate("de"), lang: "de"}, want: true},
		{name: "BaseCase10", args: args{name: Generate("de"), lang: "de"}, want: true},
		{name: "BaseCase11", args: args{name: Generate("en"), lang: "en"}, want: true},
		{name: "BaseCase12", args: args{name: Generate("en"), lang: "en"}, want: true},
		{name: "BaseCase13", args: args{name: Generate("en"), lang: "en"}, want: true},
		{name: "BaseCase14", args: args{name: Generate("en"), lang: "en"}, want: true},
		{name: "BaseCase15", args: args{name: Generate("en"), lang: "en"}, want: true},
		{name: "BaseCase16", args: args{name: Generate("en"), lang: "en"}, want: true},
		{name: "BaseCase17", args: args{name: Generate("en"), lang: "en"}, want: true},
		{name: "BaseCase18", args: args{name: Generate("en"), lang: "en"}, want: true},
		{name: "BaseCase19", args: args{name: Generate("en"), lang: "en"}, want: true},
		{name: "BaseCase20", args: args{name: Generate("en"), lang: "en"}, want: true},
		{name: "BaseCase21", args: args{name: Generate("en"), lang: "en"}, want: true},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := IsGeneratedName(tt.args.lang, tt.args.name); got != tt.want {
				t.Errorf("isGeneratedName() = %v, want %v", got, tt.want)
			}
		})
	}
}
