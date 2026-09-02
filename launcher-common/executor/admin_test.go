package executor

import (
	"crypto/x509"
	"crypto/x509/pkix"
	"io"
	"net"
	"testing"

	"github.com/luskaner/ageLANServer/common/executor/exec"
)

// fakeRunner implements Runner for tests without touching package globals.
type fakeRunner struct {
	isAdmin bool
	exec    func(o exec.Options) *exec.Result
}

func (f fakeRunner) IsAdmin() bool { return f.isAdmin }
func (f fakeRunner) Exec(o exec.Options) *exec.Result {
	if f.exec == nil {
		return &exec.Result{}
	}
	return f.exec(o)
}

func TestRunSetUp_ForwardsFlags(t *testing.T) {
	var captured exec.Options
	ex := NewExecutor(fakeRunner{exec: func(o exec.Options) *exec.Result {
		captured = o
		return &exec.Result{}
	}})

	cert := &x509.Certificate{Raw: []byte("raw"), Subject: pkix.Name{CommonName: "test"}}
	ip := net.ParseIP("1.2.3.4")
	result := ex.RunSetUp("age2", ip, true, cert, "/logs", nil, nil)
	if result == nil {
		t.Fatal("nil result")
	}
	foundGame := false
	for _, a := range captured.Args {
		if a == "--game=age2" {
			foundGame = true
		}
	}
	if !foundGame {
		t.Errorf("expected --game=age2 in args, got %v", captured.Args)
	}
	if !captured.AsAdmin {
		t.Error("AsAdmin should be true")
	}
	if captured.File == "" {
		t.Error("File empty")
	}
}

func TestRunSetUp_NilCert(t *testing.T) {
	ex := NewExecutor(fakeRunner{})
	ip := net.ParseIP("10.0.0.1")
	res := ex.RunSetUp("age2", ip, false, nil, "", nil, nil)
	if res == nil {
		t.Fatal("nil")
	}
}

func TestRunRevert_FailFastTrue(t *testing.T) {
	var captured exec.Options
	ex := NewExecutor(fakeRunner{exec: func(o exec.Options) *exec.Result {
		captured = o
		return &exec.Result{}
	}})
	ex.RunRevert(true, true, true, "/logs", nil, nil)
	foundIP, foundCert := false, false
	for _, a := range captured.Args {
		if a == "--ip" {
			foundIP = true
		}
		if a == "--localCert" {
			foundCert = true
		}
	}
	if !foundIP || !foundCert {
		t.Errorf("failfast true should set both IPs and Certs, got %v", captured.Args)
	}
}

func TestRunRevert_FailFastFalseSetsRemoveAll(t *testing.T) {
	var captured exec.Options
	ex := NewExecutor(fakeRunner{exec: func(o exec.Options) *exec.Result {
		captured = o
		return &exec.Result{}
	}})
	ex.RunRevert(false, false, false, "", nil, nil)
	foundAll := false
	for _, a := range captured.Args {
		if a == "--all" {
			foundAll = true
		}
	}
	if !foundAll {
		t.Errorf("failfast false should set --all, got %v", captured.Args)
	}
}

func TestRun_FlushCacheAgentAndFlushCache(t *testing.T) {
	ex := NewExecutor(fakeRunner{exec: func(o exec.Options) *exec.Result { return &exec.Result{Pid: 123} }})

	file, result := ex.RunFlushCacheAgent(true, false, "/logs", nil, nil)
	if file == "" {
		t.Error("file empty")
	}
	if result.Pid != 123 {
		t.Errorf("Pid=%d want 123", result.Pid)
	}
	file2, result2 := ex.RunFlushCache(false, true, "/logs", nil, nil)
	if file2 == "" {
		t.Error("file2 empty")
	}
	if result2 == nil {
		t.Fatal("nil")
	}
}

func TestRun_OptionsFnMutation(t *testing.T) {
	ex := NewExecutor(fakeRunner{exec: func(o exec.Options) *exec.Result {
		if o.File != "mutated" {
			t.Errorf("optionsFn not applied, File=%q", o.File)
		}
		return &exec.Result{}
	}})
	ex.RunSetUp("age2", nil, false, nil, "", nil, func(o *exec.Options) { o.File = "mutated" })
}

func TestRun_OutRedirectionNonWindows(t *testing.T) {
	var captured exec.Options
	ex := NewExecutor(fakeRunner{isAdmin: true, exec: func(o exec.Options) *exec.Result {
		captured = o
		return &exec.Result{}
	}})

	var buf io.Writer = &testWriter{}
	ex.RunRevert(true, false, true, "", buf, nil)
	if captured.Stdout != buf {
		t.Error("Stdout should be set when isAdmin true")
	}
}

type testWriter struct{}

func (t *testWriter) Write(p []byte) (n int, err error) { return len(p), nil }

func TestRunFlushCache_OptionsFn(t *testing.T) {
	ex := NewExecutor(fakeRunner{exec: func(o exec.Options) *exec.Result {
		if o.File != "custom" {
			t.Errorf("optionsFn File not mutated")
		}
		return &exec.Result{}
	}})
	ex.RunFlushCache(true, false, "", nil, func(o *exec.Options) { o.File = "custom" })
}

func TestDefaultExecutorIsNotNil(t *testing.T) {
	if Default == nil {
		t.Fatal("Default executor should never be nil")
	}
}
