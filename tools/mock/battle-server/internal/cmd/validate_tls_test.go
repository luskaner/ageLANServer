//go:build windows

package cmd

import "testing"

// Regression: a one-sided TLS pair (-sslCert without -sslKey or vice versa)
// passed validation but silently downgraded serving to plaintext.
func TestValidateTLSRequiresPairOrNone(t *testing.T) {
	cases := []struct {
		name    string
		cert    string
		key     string
		wantErr bool
	}{
		{"both empty is plaintext mode", "", "", false},
		{"cert only must fail", "/tmp/cert.pem", "", true},
		{"key only must fail", "", "/tmp/key.pem", true},
	}
	for _, tc := range cases {
		if err := validateTLS(tc.cert, tc.key); (err != nil) != tc.wantErr {
			t.Errorf("%s: err = %v, wantErr %v", tc.name, err, tc.wantErr)
		}
	}
}
