package require

import (
	"reflect"
	"testing"
)

func Equal[V comparable](t *testing.T, got, expected V) {
	t.Helper()

	if expected != got {
		t.Fatalf(`
|> Expected:
%v

<| Got:
%v

`, expected, got)
	}
}

func NotEqual[V comparable](t *testing.T, got, expected V) {
	t.Helper()

	if expected == got {
		t.Fatalf(`
|> Expected:
%v

<| Got:
%v

`, expected, got)
	}
}

func DeepEqual[V comparable](t *testing.T, got, expected V) {
	t.Helper()
	if !reflect.DeepEqual(expected, got) {
		t.Fatalf(`
	|> Expected:
	%v
	
	<| Got:
	%v
	
	`, expected, got)
	}
}

func AssertNoError(t *testing.T, err error) {
	t.Helper()
	if err != nil {
		t.Fatal(err)
	}
}

func AssertError(t *testing.T, err error) {
	t.Helper()
	if err == nil {
		t.Fatal("expected error")
	}
}

func AssertEqual[T comparable](t *testing.T, expected T, actual T) {
	t.Helper()
	if expected != actual {
		t.Fatalf("expected (%+v) is not equal to actual (%+v)", expected, actual)
	}
}

func AssertNotEqual[T comparable](t *testing.T, expected T, actual T) {
	t.Helper()
	if expected == actual {
		t.Fatalf("expected (%+v) is equal to actual (%+v)", expected, actual)
	}
}

func AssertSliceEqual[T comparable](t *testing.T, expected []T, actual []T) {
	t.Helper()
	if len(expected) != len(actual) {
		t.Fatalf("expected (%+v) is not equal to actual (%+v): len(expected)=%d len(actual)=%d",
			expected, actual, len(expected), len(actual))
	}
	for i := range expected {
		if expected[i] != actual[i] {
			t.Fatalf("expected[%d] (%+v) is not equal to actual[%d] (%+v)",
				i, expected[i],
				i, actual[i])
		}
	}
}
