package testing_helper

import (
	"strings"
	"testing"
)

// AssertReturnError asserts if the returned eror matched the expected error.
// If the returned error is nil but the expected is not, test fails.
// If returned error is non-nil but the expected is nil, test fails.
// If both expected and returned error are non-nil, assert if the expected
// error message is a substring of the returned erorr message.
func AssertReturnError(t *testing.T, expected error, actual error) {
	if expected != nil && actual == nil {
		t.Fatalf("expected(err): non-nil, actual(err): nil")
	}
	if expected != nil && actual != nil && !strings.Contains(strings.ToLower(actual.Error()), strings.ToLower(expected.Error())) {
		t.Fatalf("expected(err.Error()): to contain %s, actual(err.Error()): %s", expected, actual)
	}
	if expected == nil && actual != nil {
		t.Fatalf("expected(err): nil, actual(err): %s", actual)
	}
}
