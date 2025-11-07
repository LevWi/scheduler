package common

import (
	"testing"
	"time"
)

func TestGetNextMonday(t *testing.T) {
	loc := time.UTC

	testCases := []struct {
		input    time.Time
		expected time.Time
	}{
		{
			input:    time.Date(2025, 11, 3, 12, 9, 33, 0, loc),
			expected: time.Date(2025, 11, 10, 0, 0, 0, 0, loc),
		},
		{
			input:    time.Date(2025, 11, 4, 12, 9, 33, 0, loc),
			expected: time.Date(2025, 11, 10, 0, 0, 0, 0, loc),
		},
		{
			input:    time.Date(2025, 11, 5, 12, 9, 33, 0, loc),
			expected: time.Date(2025, 11, 10, 0, 0, 0, 0, loc),
		},
		{
			input:    time.Date(2025, 11, 6, 12, 9, 33, 0, loc),
			expected: time.Date(2025, 11, 10, 0, 0, 0, 0, loc),
		},
		{
			input:    time.Date(2025, 11, 7, 12, 19, 33, 0, loc),
			expected: time.Date(2025, 11, 10, 0, 0, 0, 0, loc),
		},
		{
			input:    time.Date(2025, 11, 8, 12, 39, 33, 0, loc),
			expected: time.Date(2025, 11, 10, 0, 0, 0, 0, loc),
		},
		{
			input:    time.Date(2025, 11, 9, 12, 29, 33, 0, loc),
			expected: time.Date(2025, 11, 10, 0, 0, 0, 0, loc),
		},
	}

	for _, tc := range testCases {
		t.Run(tc.input.Format(time.RFC822), func(t *testing.T) {
			got := NextMonday(tc.input)
			if !got.Equal(tc.expected) {
				t.Errorf("GetNextMonday() \n input: %s\n   got: %s\n  want: %s", tc.input.Format(time.RFC3339Nano), got.Format(time.RFC3339Nano), tc.expected.Format(time.RFC3339Nano))
			}
		})
	}
}
