package server

import (
	"testing"
	"time"
)

func TestCheckTimeParse(t *testing.T) {

	tm, err := parseTime("2024-10-06T18:33:47.072Z")
	if err != nil {
		t.Fatal(err)
	}
	t.Log(tm)

	expected := time.Date(2024, 10, 6, 18, 33, 47, 72*1000000, time.UTC)
	if expected != tm {
		t.Fatalf("expected %v, got %v", expected, tm)
	}

	tm, err = parseTime("2022-10-06T18:33:47Z")
	if err != nil {
		t.Fatal(err)
	}
	t.Log(tm)

	expected = time.Date(2022, 10, 6, 18, 33, 47, 0, time.UTC)
	if expected != tm {
		t.Fatalf("expected %v, got %v", expected, tm)
	}
}
