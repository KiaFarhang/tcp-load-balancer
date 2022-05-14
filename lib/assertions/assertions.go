package assertions

import "testing"

func AssertEqual[T comparable](t *testing.T, got, want T) {
	t.Helper()
	if got != want {
		t.Errorf("got %v expected %v", got, want)
	}
}

func AssertNoError(t *testing.T, err error) {
	if err != nil {
		t.Errorf("expected no error but got %v", err)
	}
}
