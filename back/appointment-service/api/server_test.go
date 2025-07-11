package server

import (
	common "scheduler/appointment-service/internal"
	"testing"
	"time"
)

type MockUserChecker struct {
	m map[string]struct {
		pass string
		uid  UserID
	}
}

func (uc MockUserChecker) IsExist(uid UserID) (bool, error) {
	for _, value := range uc.m {
		if uid == value.uid {
			return true, nil
		}
	}
	return false, nil
}

func (uc MockUserChecker) Check(username string, password string) (UserID, error) {
	v, ok := uc.m[username]
	if !ok {
		return "", common.ErrNotFound
	}

	if v.pass == password {
		return v.uid, nil
	}

	return "", common.ErrNotFound
}

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
