package assert

import "testing"

func Equal[T comparable](t *testing.T, got, want T) {
	t.Helper()
	if got != want {
		t.Errorf("got %v expected %v", got, want)
	}
}

func NoError(t *testing.T, err error) {
	t.Helper()
	if err != nil {
		t.Errorf("expected no error but got %v", err)
	}
}
