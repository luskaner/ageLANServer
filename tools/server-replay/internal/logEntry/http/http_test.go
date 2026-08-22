package http

import (
	"net/url"
	"testing"
	"time"

	"github.com/luskaner/ageLANServer/common/logger/serverCommunication/request"
)

func newFixture(t *testing.T, host string, port string) *Request {
	t.Helper()
	u := &url.URL{Scheme: "https", Host: host + ":" + port, Path: "/some/endpoint"}
	return NewRequest(request.Read{
		In: request.In{
			Url:    u,
			Method: "GET",
		},
	})
}

// Regression: Replay panicked on transport errors (connection refused), killing
// the whole replay run for one failed request.
func TestReplayTransportFailureDoesNotPanic(t *testing.T) {
	r := newFixture(t, "127.0.0.1", "1") // port 1: connection refused

	defer func() {
		if r := recover(); r != nil {
			t.Fatalf("panicked: %v", r)
		}
	}()
	r.Replay(nil)

	if r.response.statusCode != 0 {
		t.Fatalf("statusCode = %d, want 0 on failure", r.response.statusCode)
	}
}

// Regression: an unparsable stored URL/method used to panic via
// http.NewRequest.
func TestReplayInvalidStoredRequestDoesNotPanic(t *testing.T) {
	r := NewRequest(request.Read{
		In: request.In{
			Url:    &url.URL{Scheme: "https", Host: "127.0.0.1:443", Path: "/x"},
			Method: "BAD METHOD", // contains a space -> invalid
		},
	})

	defer func() {
		if r := recover(); r != nil {
			t.Fatalf("panicked: %v", r)
		}
	}()
	r.Replay(nil)

	if !r.ignored {
		t.Fatal("unbuildable request must be marked ignored")
	}
}

func TestUptimeExposesLoggedValue(t *testing.T) {
	r := NewRequest(request.Read{})
	r.data.Uptime.Uptime = 1234 * time.Millisecond
	if got := r.Uptime(); got != 1234*time.Millisecond {
		t.Fatalf("uptime = %v", got)
	}
}
