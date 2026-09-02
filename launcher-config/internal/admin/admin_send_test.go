package admin

import (
	"bytes"
	"encoding/gob"
	"errors"
	"net"
	"os"
	"testing"
	"time"

	"github.com/luskaner/ageLANServer/common"
	commonIpc "github.com/luskaner/ageLANServer/launcher-common/ipc"
)

// failingWriter fails on Write
type failingWriter struct{}

func (f *failingWriter) Write(p []byte) (int, error) { return 0, errors.New("write fail") }

// failingReader fails on Read
type failingReader struct{}

func (f *failingReader) Read(p []byte) (int, error)         { return 0, errors.New("read fail") }
func (f *failingReader) Write(p []byte) (int, error)        { return len(p), nil }
func (f *failingReader) Close() error                       { return nil }
func (f *failingReader) LocalAddr() net.Addr                { return nil }
func (f *failingReader) RemoteAddr() net.Addr               { return nil }
func (f *failingReader) SetDeadline(t time.Time) error      { return nil }
func (f *failingReader) SetReadDeadline(t time.Time) error  { return nil }
func (f *failingReader) SetWriteDeadline(t time.Time) error { return nil }

// statefulFailWriter fails on the failure-th Write call.
type statefulFailWriter struct {
	count  int
	failOn int
	buf    bytes.Buffer
}

func (s *statefulFailWriter) Write(p []byte) (int, error) {
	s.count++
	if s.count == s.failOn {
		return 0, errors.New("write fail on second")
	}
	return s.buf.Write(p)
}

func TestSendAgentSetupSuccessViaPipe(t *testing.T) {
	a := newTestAdmin(t)
	c1, c2 := net.Pipe()
	defer c1.Close()
	defer c2.Close()
	a.ipc = c1
	a.enc = gob.NewEncoder(c1)
	a.dec = gob.NewDecoder(c1)

	go func() {
		dec := gob.NewDecoder(c2)
		enc := gob.NewEncoder(c2)
		var typ byte
		if err := dec.Decode(&typ); err != nil {
			return
		}
		_ = enc.Encode(common.ErrSuccess)
		if typ == commonIpc.Setup {
			var cmd commonIpc.SetupCommand
			_ = dec.Decode(&cmd)
		} else {
			var cmd commonIpc.RevertCommand
			_ = dec.Decode(&cmd)
		}
		_ = enc.Encode(common.ErrSuccess)
	}()

	err, code := a.runSetUpAgent("age2", net.ParseIP("127.0.0.1"), false, []byte("cert"))
	if err != nil {
		t.Fatalf("expected nil err, got %v", err)
	}
	if code != common.ErrSuccess {
		t.Fatalf("expected success code 0, got %d", code)
	}
}

func TestSendAgentRevertSuccessViaPipe(t *testing.T) {
	a := newTestAdmin(t)
	c1, c2 := net.Pipe()
	defer c1.Close()
	defer c2.Close()
	a.ipc = c1
	a.enc = gob.NewEncoder(c1)
	a.dec = gob.NewDecoder(c1)

	go func() {
		dec := gob.NewDecoder(c2)
		enc := gob.NewEncoder(c2)
		var typ byte
		_ = dec.Decode(&typ)
		_ = enc.Encode(common.ErrSuccess)
		var cmd commonIpc.RevertCommand
		_ = dec.Decode(&cmd)
		_ = enc.Encode(common.ErrSuccess)
	}()

	err, code := a.runRevertAgent(true, false)
	if err != nil {
		t.Fatalf("expected nil err, got %v", err)
	}
	if code != common.ErrSuccess {
		t.Fatalf("code = %d, want 0", code)
	}
}

func TestSendAgentEncodeTypeFailure(t *testing.T) {
	a := newTestAdmin(t)
	a.ipc = &mockConn{}
	fw := &failingWriter{}
	a.enc = gob.NewEncoder(fw)
	a.dec = gob.NewDecoder(bytes.NewReader(nil))
	err, _ := a.sendAgent(commonIpc.Setup, "Setup", func() any { return commonIpc.SetupCommand{} })
	if err == nil {
		t.Fatal("expected encode error")
	}
}

func TestSendAgentDecodeFirstExitFailure(t *testing.T) {
	a := newTestAdmin(t)
	a.ipc = &mockConn{}
	var buf bytes.Buffer
	a.enc = gob.NewEncoder(&buf)
	a.dec = gob.NewDecoder(&failingReader{})

	err, _ := a.sendAgent(commonIpc.Setup, "Setup", func() any { return commonIpc.SetupCommand{} })
	if err == nil {
		t.Fatal("expected decode error")
	}
}

func TestSendAgentFirstDecodeNonSuccessExit(t *testing.T) {
	a := newTestAdmin(t)
	c1, c2 := net.Pipe()
	defer c1.Close()
	defer c2.Close()
	a.ipc = c1
	a.enc = gob.NewEncoder(c1)
	a.dec = gob.NewDecoder(c1)
	go func() {
		dec := gob.NewDecoder(c2)
		enc := gob.NewEncoder(c2)
		var typ byte
		_ = dec.Decode(&typ)
		_ = enc.Encode(1) // non-success
	}()
	_, code := a.runSetUpAgent("age2", nil, false, nil)
	if code != 1 {
		t.Fatalf("expected exit 1, got %d", code)
	}
}

func TestSendAgentEncodeDataFailure(t *testing.T) {
	a := newTestAdmin(t)
	a.ipc = &mockConn{}
	w := &statefulFailWriter{failOn: 2}
	a.enc = gob.NewEncoder(w)
	var buf bytes.Buffer
	_ = gob.NewEncoder(&buf).Encode(common.ErrSuccess)
	a.dec = gob.NewDecoder(&buf)
	_, _ = a.sendAgent(commonIpc.Setup, "Setup", func() any { return commonIpc.SetupCommand{} })
}

func TestSendAgentDecodeFinalFailure(t *testing.T) {
	a := newTestAdmin(t)
	c1, c2 := net.Pipe()
	defer c1.Close()
	defer c2.Close()
	a.ipc = c1
	a.enc = gob.NewEncoder(c1)
	a.dec = gob.NewDecoder(c1)
	go func() {
		dec := gob.NewDecoder(c2)
		enc := gob.NewEncoder(c2)
		var typ byte
		_ = dec.Decode(&typ)
		_ = enc.Encode(common.ErrSuccess)
		var cmd commonIpc.SetupCommand
		_ = dec.Decode(&cmd)
		c2.Close()
	}()
	_, code := a.runSetUpAgent("age2", nil, false, nil)
	_ = code
}

func TestRunSetUpAgentViaSendAgentSuccess(t *testing.T) {
	a := newTestAdmin(t)
	c1, c2 := net.Pipe()
	defer c1.Close()
	defer c2.Close()
	a.ipc = c1
	a.enc = gob.NewEncoder(c1)
	a.dec = gob.NewDecoder(c1)
	go func() {
		dec := gob.NewDecoder(c2)
		enc := gob.NewEncoder(c2)
		var typ byte
		_ = dec.Decode(&typ)
		if typ != commonIpc.Setup {
			t.Errorf("expected Setup type %d, got %d", commonIpc.Setup, typ)
		}
		_ = enc.Encode(common.ErrSuccess)
		var cmd commonIpc.SetupCommand
		_ = dec.Decode(&cmd)
		if cmd.GameId != "age3" {
			t.Errorf("gameId = %q", cmd.GameId)
		}
		_ = enc.Encode(common.ErrSuccess)
	}()
	err, code := a.runSetUpAgent("age3", net.ParseIP("1.2.3.4"), true, []byte("certdata"))
	if err != nil || code != common.ErrSuccess {
		t.Fatalf("expected success, got err %v code %d", err, code)
	}
}

func TestStopAgentIfNeededWithIPC(t *testing.T) {
	a := newTestAdmin(t)
	c1, c2 := net.Pipe()
	defer c2.Close()
	a.ipc = c1
	a.enc = gob.NewEncoder(c1)
	a.dec = gob.NewDecoder(c1)
	go func() {
		dec := gob.NewDecoder(c2)
		var cmd byte
		_ = dec.Decode(&cmd)
	}()
	a.deps.process = func(string) (string, *os.Process, error) { return "", nil, nil }
	a.deps.nativeFileName = func(bool, string) string { return "dummy.exe" }
	if !a.StopAgentIfNeeded() {
		t.Fatal("expected true when stopping via IPC with no process")
	}
	if a.ipc != nil {
		t.Fatal("ipc should be cleared after stop")
	}
}
