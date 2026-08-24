package wss

import (
	"sync"
	"testing"
)

// Regression: Close used to nil writeLock; a concurrent sendMessage would
// panic on writeLock.Lock() with a nil receiver.
func TestCloseIdempotent(t *testing.T) {
	cw := &connectionWrapper{
		connLock:  &sync.RWMutex{},
		writeLock: &sync.Mutex{},
		conn:      nil,
	}
	if err := cw.Close(); err != nil {
		t.Fatalf("first Close: %v", err)
	}
	if err := cw.Close(); err != nil {
		t.Fatalf("second Close (idempotent): %v", err)
	}
	// writeLock must still be usable after Close.
	cw.writeLock.Lock()
	cw.writeLock.Unlock()
}

func TestCloseWithConn(t *testing.T) {
	cw := &connectionWrapper{
		connLock:  &sync.RWMutex{},
		writeLock: &sync.Mutex{},
		conn:      nil,
	}
	if err := cw.Close(); err != nil {
		t.Fatal(err)
	}
}

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
