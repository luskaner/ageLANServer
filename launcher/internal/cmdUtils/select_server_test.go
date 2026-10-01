package cmdUtils

import (
	"bytes"
	"io"
	"net"
	"os"
	"strings"
	"testing"
	"time"

	"github.com/luskaner/ageLANServer/launcher/internal/server"
)

func testProcessedServers(count int) []*processedServer {
	procServers := make([]*processedServer, count)
	for i := range procServers {
		procServers[i] = &processedServer{
			MesuredIpAddress: serverMesuredIpAddress(i),
			description:      strings.Repeat("x", i+1),
		}
	}
	return procServers
}

func TestSelectServerIndexAutoSelectSingle(t *testing.T) {
	if got := selectServerIndex(testProcessedServers(1), true, strings.NewReader("")); got != 0 {
		t.Fatalf("auto-select single = %d, want 0", got)
	}
}

func TestSelectServerIndexValidInput(t *testing.T) {
	for _, tc := range []struct {
		input string
		want  int
	}{
		{"1\n", 0},
		{"2\n", 1},
		{"3\n", 2},
	} {
		reader := strings.NewReader(tc.input)
		if got := selectServerIndex(testProcessedServers(3), false, reader); got != tc.want {
			t.Errorf("input %q: index = %d, want %d", tc.input, got, tc.want)
		}
	}
}

// Regression: out-of-range input used to share the same `continue` as read
// errors; now it retries but a broken reader must give up immediately.
func TestSelectServerIndexInvalidThenValidRetries(t *testing.T) {
	reader := strings.NewReader("99\n0\n2\n")
	if got := selectServerIndex(testProcessedServers(3), false, reader); got != 1 {
		t.Fatalf("index = %d, want 1 (0-based after retry)", got)
	}
}

// Regression: on stdin EOF the old loop printed the list forever because it
// `continue`d on every Scan error. It must return -1 so the caller can fall
// back to starting its own server.
func TestSelectServerIndexEOFReturnsNegative(t *testing.T) {
	if got := selectServerIndex(testProcessedServers(3), false, strings.NewReader("")); got != -1 {
		t.Fatalf("EOF = %d, want -1", got)
	}
}

// Regression: the selection prompt lost the numbered list of discovered servers
// (name/IP/latency/...), leaving only the bare "Enter the number of..." prompt.
// Every row must be printed, numbered from 1, on each attempt.
func TestSelectServerIndexPrintsServerList(t *testing.T) {
	out := captureStdout(t, func() {
		if got := selectServerIndex(testProcessedServers(3), false, strings.NewReader("2\n")); got != 1 {
			t.Errorf("index = %d, want 1", got)
		}
	})
	for _, want := range []string{
		"Found the following 'server's:",
		"1. x\n",
		"2. xx\n",
		"3. xxx\n",
		"Enter the number of the 'server' (1-3): ",
	} {
		if !strings.Contains(out, want) {
			t.Errorf("output missing %q, got:\n%s", want, out)
		}
	}
}

func captureStdout(t *testing.T, fn func()) string {
	t.Helper()
	r, w, err := os.Pipe()
	if err != nil {
		t.Fatalf("pipe: %v", err)
	}
	orig := os.Stdout
	os.Stdout = w
	defer func() { os.Stdout = orig }()

	done := make(chan string, 1)
	go func() {
		var buf bytes.Buffer
		_, _ = io.Copy(&buf, r)
		done <- buf.String()
	}()

	fn()

	if err := w.Close(); err != nil {
		t.Fatalf("close pipe: %v", err)
	}
	return <-done
}

func serverMesuredIpAddress(i int) (res server.MesuredIpAddress) {
	res.Ip = net.IPv4(192, 168, 1, byte(10+i))
	res.Latency = time.Duration(i+1) * time.Millisecond
	return
}
