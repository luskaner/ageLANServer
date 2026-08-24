package wss

import (
	"testing"
)

// Regression: parseMessage used unchecked type assertions that panicked on
// malformed JSON from unauthenticated clients.
func TestParseMessageMalformedInput(t *testing.T) {
	cases := []struct {
		name string
		msg  map[string]any
	}{
		{"missing operation", map[string]any{}},
		{"operation is string", map[string]any{"operation": "zero"}},
		{"operation is null", map[string]any{"operation": nil}},
		{"sessionToken is number", map[string]any{"operation": float64(0), "sessionToken": float64(42)}},
		{"sessionToken is null", map[string]any{"operation": float64(0), "sessionToken": nil}},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			defer func() {
				if r := recover(); r != nil {
					t.Fatalf("panicked: %v", r)
				}
			}()
			op, sess := parseMessage(nil, tc.msg, nil)
			_ = op
			_ = sess
		})
	}
}

// Note: TestParseMessageValidLogin cannot be tested in isolation because
// parseMessage calls sessions.GetById which requires real model infrastructure.
// Full integration testing of the WSS login flow requires the server running.

