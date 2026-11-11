package api

import (
	"testing"
	"time"
)

func TestNextDate(t *testing.T) {
	now := time.Date(2024, 1, 26, 0, 0, 0, 0, time.UTC)
	tests := []struct {
		date, repeat, want string
		wantErr            bool
	}{
		// ошибки формата
		{"20240126", "", "", true},
		{"20240126", "k 34", "", true},
		{"ooops", "y", "", true},
		{"15000156", "y", "", true},
		{"20240113", "d", "", true},
		{"20240320", "d 401", "", true},
		{"20240320", "d 0", "", true},
		{"20240120", "m 40,11,19", "", true},
		{"20240222", "m -2,-3", "", true},
		{"20240222", "m 5 1,13", "", true},
		{"20230226", "w 8,4,5", "", true},
		// базовые правила
		{"16890220", "y", "20240220", false},
		{"20240229", "y", "20250301", false},
		{"20250701", "y", "20260701", false},
		{"20240113", "d 7", "20240127", false},
		{"20231225", "d 12", "20240130", false},
		{"20240228", "d 1", "20240229", false},
		// дни недели
		{"20240125", "w 1,2,3", "20240129", false},
		{"20240126", "w 7", "20240128", false},
		{"20230126", "w 4,5", "20240201", false},
		// дни месяца
		{"20231106", "m 13", "20240213", false},
		{"20240116", "m 16,5", "20240205", false},
		{"20240126", "m 25,26,7", "20240207", false},
		{"20240409", "m 31", "20240531", false},
		{"20240329", "m 10,17 12,8,1", "20240810", false},
		{"20230311", "m 07,19 05,6", "20240507", false},
		{"20230311", "m 1 1,2", "20240201", false},
		{"20240127", "m -1", "20240131", false},
		{"20240222", "m -2", "20240228", false},
		{"20240326", "m -1,-2", "20240330", false},
		{"20240201", "m -1,18", "20240218", false},
	}
	for _, tc := range tests {
		got, err := NextDate(now, tc.date, tc.repeat)
		if tc.wantErr {
			if err == nil {
				t.Errorf("NextDate(%q, %q) = %q, ожидалась ошибка", tc.date, tc.repeat, got)
			}
			continue
		}
		if err != nil {
			t.Errorf("NextDate(%q, %q): неожиданная ошибка %v", tc.date, tc.repeat, err)
			continue
		}
		if got != tc.want {
			t.Errorf("NextDate(%q, %q) = %q, ожидалось %q", tc.date, tc.repeat, got, tc.want)
		}
	}
}
