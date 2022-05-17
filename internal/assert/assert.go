/*
Package assert implements simple assertion functions for use in tests.

TODO: in reality we'd probably want a library like testify, but this is fine
for a proof of concept.
*/
package assert

import "testing"

// Equal compares two comparables of the same type and calls t.Errorf if
// they are not equal.
func Equal[T comparable](t *testing.T, got, want T) {
	t.Helper()
	if got != want {
		t.Errorf("got %v expected %v", got, want)
	}
}

// NoError calls t.Errorf if the err passed is not nil.
func NoError(t *testing.T, err error) {
	t.Helper()
	if err != nil {
		t.Errorf("expected no error but got %v", err)
	}
}
