package csvx

import (
	"errors"
	"strings"
	"testing"
)

func TestFieldError_UnwrapAndMessage(t *testing.T) {
	inner := errors.New("parse int")
	err := &FieldError{Row: 3, Column: "age", Index: 2, Err: inner}
	if !errors.Is(err, inner) {
		t.Fatal("expected Unwrap to inner")
	}
	msg := err.Error()
	if !strings.Contains(msg, "row 3") || !strings.Contains(msg, "age") {
		t.Fatalf("message %q missing context", msg)
	}
}
