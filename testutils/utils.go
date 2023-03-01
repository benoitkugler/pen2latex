package testutils

import "testing"

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
