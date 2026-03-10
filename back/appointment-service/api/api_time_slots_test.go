package api

import (
	"net/url"
	slotsdb "scheduler/appointment-service/internal/dbase/backend/slots"
	"testing"
	"time"
)

func TestGetSlotChunkFromURL(t *testing.T) {
	defaults := slotsdb.BusinessSlotSettings{DefaultChunk: 20 * time.Minute, MaxChunk: 45 * time.Minute}

	tests := []struct {
		name      string
		query     string
		expected  time.Duration
		wantError bool
	}{
		{name: "default from db", query: "", expected: 20 * time.Minute},
		{name: "custom", query: "chunk_minutes=30", expected: 30 * time.Minute},
		{name: "too small", query: "chunk_minutes=1", wantError: true},
		{name: "too big", query: "chunk_minutes=60", wantError: true},
		{name: "invalid", query: "chunk_minutes=abc", wantError: true},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			values, _ := url.ParseQuery(tc.query)
			got, err := getSlotChunkFromURL(values, defaults)
			if tc.wantError {
				if err == nil {
					t.Fatalf("expected error, got none")
				}
				return
			}
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			if got != tc.expected {
				t.Fatalf("expected %v, got %v", tc.expected, got)
			}
		})
	}
}
