package testutils

import (
	"reflect"
	"testing"
)

func Assert(t *testing.T, cond bool) {
	t.Helper()

	if !cond {
		t.Fatal("assertion error")
	}
}

func AssertNoErr(t *testing.T, err error) {
	t.Helper()

	if err != nil {
		t.Fatalf("unexpected error %v", err)
	}
}

func AssertEqual(t *testing.T, got, exp any) {
	t.Helper()
	if !reflect.DeepEqual(exp, got) {
		t.Fatalf("expected %v, got %v", exp, got)
	}
}
