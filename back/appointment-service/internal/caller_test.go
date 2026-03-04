package common

import (
	"strings"
	"testing"
)

func TestCallerName(t *testing.T) {
	name := CallerName(0)

	if name == UnknownFn {
		t.Fatalf("expected caller name, got %q", name)
	}

	if !strings.HasSuffix(name, ".TestCallerName") {
		t.Fatalf("unexpected caller name: %s", name)
	}

	name = CallerShort(0)
	if name != "internal.TestCallerName" {
		t.Fatalf("unexpected caller name: %s", name)
	}

	name = dummy{}.foo()
	if !strings.HasSuffix(name, ".foo") {
		t.Fatalf("unexpected caller name: %s", name)
	}

	name = dummy{}.bar()
	if name != "bar" {
		t.Fatalf("unexpected caller name: %s", name)
	}
}

type dummy struct{}

func (dummy) foo() string {
	return CallerName(0)
}

func (dummy) bar() string {
	return CallerFuncOnly(0)
}
