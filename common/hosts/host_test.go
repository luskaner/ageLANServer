package hosts

import "testing"

func TestHostString(t *testing.T) {
	cases := map[string]string{
		"example.com":     "example.com",
		"sub.example.com": "sub.example.com",
		// Unicode host is converted to its ASCII (punycode) form.
		"münchen.example": "xn--mnchen-3ya.example",
	}
	for in, want := range cases {
		if got := Host(in).String(); got != want {
			t.Errorf("Host(%q).String() = %q, want %q", in, got, want)
		}
	}
}
