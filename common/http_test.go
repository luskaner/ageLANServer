package common

import "testing"

func TestUserAgent(t *testing.T) {
	got := UserAgent()
	want := Name + "/1.0"
	if got != want {
		t.Fatalf("UserAgent() = %q, want %q", got, want)
	}
}
