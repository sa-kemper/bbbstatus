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

package StatsAggregator

import (
	"reflect"
	"testing"
	"time"
)

func Test_generateCalendarWeeks(t *testing.T) {
	berlin, _ := time.LoadLocation("Europe/Berlin")
	type args struct {
		year int
	}
	tests := []struct {
		name       string
		args       args
		wantResult map[int]TimeFrame
	}{
		{name: "Basecase", args: args{year: 2025}, wantResult: map[int]TimeFrame{ // the function seems to work, but the unit tests seems broken af.
			1:  {Start: time.Date(2024, 12, 30, 0, 0, 0, 0, berlin), End: time.Date(2025, 1, 05, 23, 59, 59, 0, berlin)},
			2:  {Start: time.Date(2025, 01, 06, 0, 0, 0, 0, berlin), End: time.Date(2025, 1, 12, 23, 59, 59, 0, berlin)},
			3:  {Start: time.Date(2025, 01, 13, 0, 0, 0, 0, berlin), End: time.Date(2025, 1, 19, 23, 59, 59, 0, berlin)},
			4:  {Start: time.Date(2025, 01, 20, 0, 0, 0, 0, berlin), End: time.Date(2025, 1, 26, 23, 59, 59, 0, berlin)},
			5:  {Start: time.Date(2025, 01, 27, 0, 0, 0, 0, berlin), End: time.Date(2025, 2, 02, 23, 59, 59, 0, berlin)},
			6:  {Start: time.Date(2025, 02, 03, 0, 0, 0, 0, berlin), End: time.Date(2025, 02, 9, 23, 59, 59, 0, berlin)},
			7:  {Start: time.Date(2025, 02, 10, 0, 0, 0, 0, berlin), End: time.Date(2025, 02, 16, 23, 59, 59, 0, berlin)},
			8:  {Start: time.Date(2025, 02, 17, 0, 0, 0, 0, berlin), End: time.Date(2025, 02, 23, 23, 59, 59, 0, berlin)},
			9:  {Start: time.Date(2025, 02, 24, 0, 0, 0, 0, berlin), End: time.Date(2025, 03, 02, 23, 59, 59, 0, berlin)},
			10: {Start: time.Date(2025, 03, 03, 0, 0, 0, 0, berlin), End: time.Date(2025, 03, 9, 23, 59, 59, 0, berlin)},
			11: {Start: time.Date(2025, 03, 10, 0, 0, 0, 0, berlin), End: time.Date(2025, 03, 16, 23, 59, 59, 0, berlin)},
			12: {Start: time.Date(2025, 03, 17, 0, 0, 0, 0, berlin), End: time.Date(2025, 03, 23, 23, 59, 59, 0, berlin)},
			13: {Start: time.Date(2025, 03, 24, 0, 0, 0, 0, berlin), End: time.Date(2025, 3, 30, 23, 59, 59, 0, berlin)},
			14: {Start: time.Date(2025, 03, 31, 0, 0, 0, 0, berlin), End: time.Date(2025, 4, 06, 23, 59, 59, 0, berlin)},
			15: {Start: time.Date(2025, 04, 07, 0, 0, 0, 0, berlin), End: time.Date(2025, 4, 13, 23, 59, 59, 0, berlin)},
			16: {Start: time.Date(2025, 04, 14, 0, 0, 0, 0, berlin), End: time.Date(2025, 4, 20, 23, 59, 59, 0, berlin)},
			17: {Start: time.Date(2025, 04, 21, 0, 0, 0, 0, berlin), End: time.Date(2025, 4, 27, 23, 59, 59, 0, berlin)},
			18: {Start: time.Date(2025, 04, 28, 0, 0, 0, 0, berlin), End: time.Date(2025, 5, 04, 23, 59, 59, 0, berlin)},
			19: {Start: time.Date(2025, 05, 05, 0, 0, 0, 0, berlin), End: time.Date(2025, 5, 11, 23, 59, 59, 0, berlin)},
			20: {Start: time.Date(2025, 05, 12, 0, 0, 0, 0, berlin), End: time.Date(2025, 5, 18, 23, 59, 59, 0, berlin)},
			21: {Start: time.Date(2025, 05, 19, 0, 0, 0, 0, berlin), End: time.Date(2025, 5, 25, 23, 59, 59, 0, berlin)},
			22: {Start: time.Date(2025, 05, 26, 0, 0, 0, 0, berlin), End: time.Date(2025, 6, 01, 23, 59, 59, 0, berlin)},
			23: {Start: time.Date(2025, 06, 02, 0, 0, 0, 0, berlin), End: time.Date(2025, 6, 8, 23, 59, 59, 0, berlin)},
			24: {Start: time.Date(2025, 06, 9, 0, 0, 0, 0, berlin), End: time.Date(2025, 6, 15, 23, 59, 59, 0, berlin)},
			25: {Start: time.Date(2025, 06, 16, 0, 0, 0, 0, berlin), End: time.Date(2025, 6, 22, 23, 59, 59, 0, berlin)},
			26: {Start: time.Date(2025, 06, 23, 0, 0, 0, 0, berlin), End: time.Date(2025, 6, 29, 23, 59, 59, 0, berlin)},
			27: {Start: time.Date(2025, 6, 30, 0, 0, 0, 0, berlin), End: time.Date(2025, 7, 06, 23, 59, 59, 0, berlin)},
			28: {Start: time.Date(2025, 7, 07, 0, 0, 0, 0, berlin), End: time.Date(2025, 7, 13, 23, 59, 59, 0, berlin)},
			29: {Start: time.Date(2025, 7, 14, 0, 0, 0, 0, berlin), End: time.Date(2025, 7, 20, 23, 59, 59, 0, berlin)},
			30: {Start: time.Date(2025, 7, 21, 0, 0, 0, 0, berlin), End: time.Date(2025, 7, 27, 23, 59, 59, 0, berlin)},
			31: {Start: time.Date(2025, 7, 28, 0, 0, 0, 0, berlin), End: time.Date(2025, 8, 03, 23, 59, 59, 0, berlin)},
			32: {Start: time.Date(2025, 8, 04, 0, 0, 0, 0, berlin), End: time.Date(2025, 8, 10, 23, 59, 59, 0, berlin)},
			33: {Start: time.Date(2025, 8, 11, 0, 0, 0, 0, berlin), End: time.Date(2025, 8, 17, 23, 59, 59, 0, berlin)},
			34: {Start: time.Date(2025, 8, 18, 0, 0, 0, 0, berlin), End: time.Date(2025, 8, 24, 23, 59, 59, 0, berlin)},
			35: {Start: time.Date(2025, 8, 25, 0, 0, 0, 0, berlin), End: time.Date(2025, 8, 31, 23, 59, 59, 0, berlin)},
			36: {Start: time.Date(2025, 9, 01, 0, 0, 0, 0, berlin), End: time.Date(2025, 9, 07, 23, 59, 59, 0, berlin)},
			37: {Start: time.Date(2025, 9, 8, 0, 0, 0, 0, berlin), End: time.Date(2025, 9, 14, 23, 59, 59, 0, berlin)},
			38: {Start: time.Date(2025, 9, 15, 0, 0, 0, 0, berlin), End: time.Date(2025, 9, 21, 23, 59, 59, 0, berlin)},
			39: {Start: time.Date(2025, 9, 22, 0, 0, 0, 0, berlin), End: time.Date(2025, 9, 28, 23, 59, 59, 0, berlin)},
			40: {Start: time.Date(2025, 9, 29, 0, 0, 0, 0, berlin), End: time.Date(2025, 10, 05, 23, 59, 59, 0, berlin)},
			41: {Start: time.Date(2025, 10, 06, 0, 0, 0, 0, berlin), End: time.Date(2025, 10, 12, 23, 59, 59, 0, berlin)},
			42: {Start: time.Date(2025, 10, 13, 0, 0, 0, 0, berlin), End: time.Date(2025, 10, 19, 23, 59, 59, 0, berlin)},
			43: {Start: time.Date(2025, 10, 20, 0, 0, 0, 0, berlin), End: time.Date(2025, 10, 26, 23, 59, 59, 0, berlin)},
			44: {Start: time.Date(2025, 10, 27, 0, 0, 0, 0, berlin), End: time.Date(2025, 11, 02, 23, 59, 59, 0, berlin)},
			45: {Start: time.Date(2025, 11, 03, 0, 0, 0, 0, berlin), End: time.Date(2025, 11, 9, 23, 59, 59, 0, berlin)},
			46: {Start: time.Date(2025, 11, 10, 0, 0, 0, 0, berlin), End: time.Date(2025, 11, 16, 23, 59, 59, 0, berlin)},
			47: {Start: time.Date(2025, 11, 17, 0, 0, 0, 0, berlin), End: time.Date(2025, 11, 23, 23, 59, 59, 0, berlin)},
			48: {Start: time.Date(2025, 11, 24, 0, 0, 0, 0, berlin), End: time.Date(2025, 11, 30, 23, 59, 59, 0, berlin)},
			49: {Start: time.Date(2025, 12, 01, 0, 0, 0, 0, berlin), End: time.Date(2025, 12, 07, 23, 59, 59, 0, berlin)},
			50: {Start: time.Date(2025, 12, 8, 0, 0, 0, 0, berlin), End: time.Date(2025, 12, 14, 23, 59, 59, 0, berlin)},
			51: {Start: time.Date(2025, 12, 15, 0, 0, 0, 0, berlin), End: time.Date(2025, 12, 21, 23, 59, 59, 0, berlin)},
			52: {Start: time.Date(2025, 12, 22, 0, 0, 0, 0, berlin), End: time.Date(2025, 12, 28, 23, 59, 59, 0, berlin)},
		}},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if gotResult := generateCalendarWeeks(tt.args.year, berlin); !reflect.DeepEqual(gotResult, tt.wantResult) {
				// This complex unit test needs some further investigation
				for i := 1; i < 53; i++ {
					_, gotWeek := gotResult[i].Start.ISOWeek()
					_, wantWeek := tt.wantResult[i].Start.ISOWeek()
					if gotWeek != wantWeek {
						t.Errorf("generateCalendarWeeks(): week %d is not equally set Got:%d want: %d", i, gotWeek, wantWeek)
					}
					if !gotResult[i].Start.Equal(tt.wantResult[i].Start) {
						t.Errorf("generateCalendarWeeks(): week %d start is not equally set Got:%s want: %s", i, gotResult[i].Start.Format("Monday the 02th day of January at 15:04:05"), tt.wantResult[i].Start.Format("Monday the 02th day of January at 15:04:05"))
					}
					gotEnd := gotResult[i].End
					wantEnd := tt.wantResult[i].End
					if !gotEnd.Equal(wantEnd) {
						t.Errorf("generateCalendarWeeks(): week %d end is not equally set Got:%s want: %s", i, gotEnd.Format("Monday the 02th day of January at 15:04:05"), wantEnd.Format("Monday the 02th day of January at 15:04:05"))
					}

				}
				t.Errorf("generateCalendarWeeks() = %v\nWANT = %v", gotResult, tt.wantResult)
			}
		})
	}
}
