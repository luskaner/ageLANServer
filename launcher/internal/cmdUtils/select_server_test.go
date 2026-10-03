package cmdUtils

import (
	"bytes"
	"io"
	"net"
	"os"
	"strings"
	"testing"
	"time"

	"github.com/luskaner/ageLANServer/launcher/internal/dialog"
	"github.com/luskaner/ageLANServer/launcher/internal/server"
)

func testProcessedServers(count int) []*processedServer {
	procServers := make([]*processedServer, count)
	for i := range procServers {
		procServers[i] = &processedServer{
			MesuredIpAddress: serverMesuredIpAddress(i),
			description:      strings.Repeat("x", i+1),
			label:            strings.Repeat("y", i+1),
		}
	}
	return procServers
}

// fakeDialog records what the caller handed to the active dialog and returns a
// canned answer.
type fakeDialog struct {
	index int
	ok    bool

	gotCandidates []dialog.ServerCandidate
	gotStdin      io.Reader

	gotListed    []dialog.ServerCandidate
	listCalled   int
	selectCalled int
}

func (f *fakeDialog) Name() string { return "fake" }

func (f *fakeDialog) SelectServer(servers []dialog.ServerCandidate, stdin io.Reader) (int, bool) {
	f.selectCalled++
	f.gotCandidates = servers
	f.gotStdin = stdin
	return f.index, f.ok
}

func (f *fakeDialog) ListCandidates(servers []dialog.ServerCandidate) {
	f.listCalled++
	f.gotListed = servers
}

func (f *fakeDialog) ConfirmStartServer(string, io.Reader) bool { return true }

func installFakeDialog(t *testing.T, d dialog.Dialog) {
	t.Helper()
	dialog.Reset()
	dialog.Set(d)
	t.Cleanup(dialog.Reset)
}

// Auto-select answers on its own, but it still lists the candidate, so it needs
// the console backend rather than whatever a previous test left installed.
func TestSelectDiscoveredServerAutoSelectSingle(t *testing.T) {
	dialog.Reset()
	t.Cleanup(dialog.Reset)
	if got, ok := selectDiscoveredServer(testProcessedServers(1), true, strings.NewReader("")); !ok || got != 0 {
		t.Fatalf("auto-select single = (%d, %t), want (0, true)", got, ok)
	}
}

func TestSelectDiscoveredServerAutoSelectIgnoredWhenMultiple(t *testing.T) {
	fake := &fakeDialog{index: 1, ok: true}
	installFakeDialog(t, fake)
	got, ok := selectDiscoveredServer(testProcessedServers(3), true, strings.NewReader("2\n"))
	if !ok || got != 1 {
		t.Fatalf("= (%d, %t), want (1, true)", got, ok)
	}
	if len(fake.gotCandidates) != 3 {
		t.Fatalf("the dialog got %d descriptions, want 3", len(fake.gotCandidates))
	}
}

func TestSelectDiscoveredServerDelegatesDescriptionsInOrder(t *testing.T) {
	procServers := testProcessedServers(3)
	procServers[0].description = "first"
	procServers[1].description = "second"
	procServers[2].description = "third"
	fake := &fakeDialog{index: 2, ok: true}
	installFakeDialog(t, fake)
	got, ok := selectDiscoveredServer(procServers, false, strings.NewReader(""))
	if !ok || got != 2 {
		t.Fatalf("= (%d, %t), want (2, true)", got, ok)
	}
	want := []string{"first", "second", "third"}
	for i := range want {
		if fake.gotCandidates[i].Description != want[i] {
			t.Fatalf("descriptions = %v, want %v (latency order must be preserved)", fake.gotCandidates, want)
		}
	}
}

// Regression: the compact label exists because the graphical list has a fixed
// 241 px of width on Windows. Both forms must reach the backend, or the narrow
// one loses detail and the console one loses the address.
func TestSelectDiscoveredServerPassesBothDescriptionAndLabel(t *testing.T) {
	procServers := testProcessedServers(2)
	procServers[0].description = "192.168.1.10, 192.168.1.11 (a.local, b.local) - 5 ms (v1.11.0)"
	procServers[0].label = "192.168.1.10 - 5 ms (v1.11.0)"
	procServers[1].description = "10.0.0.1 (c.local) - 1 ms (v1.11.0)"
	procServers[1].label = "10.0.0.1 - 1 ms (v1.11.0)"
	fake := &fakeDialog{index: 0, ok: true}
	installFakeDialog(t, fake)
	if _, ok := selectDiscoveredServer(procServers, false, strings.NewReader("")); !ok {
		t.Fatal("SelectServer was not called")
	}
	if got := fake.gotCandidates[0]; got.Description != procServers[0].description || got.Label != procServers[0].label {
		t.Fatalf("candidate 0 = %+v, want description %q and label %q",
			got, procServers[0].description, procServers[0].label)
	}
	// The compact label must be what a narrow GUI can show.
	if len(fake.gotCandidates[0].Label) >= len(fake.gotCandidates[0].Description) {
		t.Errorf("label %q is not shorter than description %q, it would clip anyway",
			fake.gotCandidates[0].Label, fake.gotCandidates[0].Description)
	}
}

// Regression: auto-select must not depend on any backend being installed, it
// never asks anything.
func TestSelectDiscoveredServerAutoSelectNeedsNoPrompt(t *testing.T) {
	fake := &fakeDialog{ok: false}
	installFakeDialog(t, fake)
	if got, ok := selectDiscoveredServer(testProcessedServers(1), true, strings.NewReader("")); !ok || got != 0 {
		t.Fatalf("= (%d, %t), want (0, true)", got, ok)
	}
	if fake.selectCalled != 0 {
		t.Errorf("auto-select asked the dialog anyway (%d SelectServer calls)", fake.selectCalled)
	}
}

// Regression: auto-select must still list the candidate before announcing it,
// otherwise the console user loses the "Found the following servers:" block
// that has always been printed on this path.
func TestSelectDiscoveredServerAutoSelectListsCandidate(t *testing.T) {
	fake := &fakeDialog{}
	installFakeDialog(t, fake)
	procServers := testProcessedServers(1)
	procServers[0].description = "only one"
	if got, ok := selectDiscoveredServer(procServers, true, strings.NewReader("")); !ok || got != 0 {
		t.Fatalf("= (%d, %t), want (0, true)", got, ok)
	}
	if fake.listCalled != 1 {
		t.Fatalf("ListCandidates called %d times, want 1", fake.listCalled)
	}
	if len(fake.gotListed) != 1 || fake.gotListed[0].Description != "only one" {
		t.Fatalf("listed %v, want the candidate with description [only one]", fake.gotListed)
	}
}

func TestSelectDiscoveredServerCancelledReturnsNotOK(t *testing.T) {
	installFakeDialog(t, &fakeDialog{ok: false})
	if got, ok := selectDiscoveredServer(testProcessedServers(3), false, strings.NewReader("1\n")); ok {
		t.Fatalf("= (%d, true), want ok=false so the caller starts its own server", got)
	}
}

// A backend answering ok=true with an out-of-range index must not be indexed;
// the caller guards against it.
func TestSelectDiscoveredServerOutOfRangeIndexIsNotFatal(t *testing.T) {
	installFakeDialog(t, &fakeDialog{index: 99, ok: true})
	got, ok := selectDiscoveredServer(testProcessedServers(3), false, strings.NewReader("1\n"))
	if !ok || got != 99 {
		t.Fatalf("= (%d, %t), want the backend answer passed through", got, ok)
	}
}

// Regression: DiscoverServersAndSelectBestIpAddr indexes procServers with the
// answer coming from the backend, so an out-of-range answer must be ignored
// instead of panicking.
func TestUsableServerIndexRejectsOutOfRangeIndex(t *testing.T) {
	for _, tc := range []struct {
		idx  int
		ok   bool
		want int
	}{
		{0, true, 0},
		{2, true, 2},
		{3, true, -1}, // == procCount
		{99, true, -1},
		{-1, true, -1},
		{1, false, -1}, // declined
	} {
		if got := usableServerIndex(tc.idx, tc.ok, 3); got != tc.want {
			t.Errorf("usableServerIndex(%d, %t, 3) = %d, want %d", tc.idx, tc.ok, got, tc.want)
		}
	}
}

func TestSelectDiscoveredServerFallsBackToConsoleByDefault(t *testing.T) {
	dialog.Reset()
	got, ok := selectDiscoveredServer(testProcessedServers(3), false, strings.NewReader("2\n"))
	if !ok || got != 1 {
		t.Fatalf("= (%d, %t), want (1, true)", got, ok)
	}
}

// Regression: with Dialog='false' the auto-select output must stay byte for
// byte what it was before graphical dialogs existed: the numbered list first,
// then the auto-select announcement, and no prompt.
func TestSelectDiscoveredServerAutoSelectConsoleOutputUnchanged(t *testing.T) {
	dialog.Reset()
	procServers := testProcessedServers(1)
	procServers[0].description = "solo"
	got := captureStdout(t, func() {
		var idx int
		var selected bool
		idx, selected = selectDiscoveredServer(procServers, true, strings.NewReader(""))
		if !selected || idx != 0 {
			t.Errorf("= (%d, %t), want (0, true)", idx, selected)
		}
	})
	want := "Found the following servers:\n1. solo\nAuto-selecting the only found server.\n"
	if got != want {
		t.Fatalf("output =\n%q\nwant\n%q", got, want)
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
