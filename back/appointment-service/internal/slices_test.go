package common

import (
	"errors"
	"reflect"
	"strconv"
	"testing"
)

func TestMap_Success(t *testing.T) {
	input := []int{1, 2, 3}

	result, err := MapE(input, func(v int) (string, error) {
		return strconv.Itoa(v), nil
	})

	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	expected := []string{"1", "2", "3"}

	if !reflect.DeepEqual(result, expected) {
		t.Fatalf("expected %v, got %v", expected, result)
	}
}

func TestMap_Error(t *testing.T) {
	input := []int{1, 2, 3}

	expectedErr := errors.New("boom")

	_, err := MapE(input, func(v int) (string, error) {
		if v == 2 {
			return "", expectedErr
		}
		return strconv.Itoa(v), nil
	})

	if err == nil {
		t.Fatal("expected error but got nil")
	}

	if !errors.Is(err, expectedErr) {
		t.Fatalf("expected %v, got %v", expectedErr, err)
	}
}

func TestMap_EmptySlice(t *testing.T) {
	var input []int

	result, err := MapE(input, func(v int) (string, error) {
		return strconv.Itoa(v), nil
	})

	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if len(result) != 0 {
		t.Fatalf("expected empty slice, got %v", result)
	}
}
