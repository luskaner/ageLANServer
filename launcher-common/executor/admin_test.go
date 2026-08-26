package executor

import (
	"crypto/x509"
	"crypto/x509/pkix"
	"io"
	"net"
	"testing"

	"github.com/luskaner/ageLANServer/common/executor/exec"
)

func TestRunSetUp_ForwardsFlags(t *testing.T) {
	orig := execFn
	var captured exec.Options
	execFn = func(o exec.Options) *exec.Result {
		captured = o
		return &exec.Result{}
	}
	defer func() { execFn = orig }()

	cert := &x509.Certificate{Raw: []byte("raw"), Subject: pkix.Name{CommonName: "test"}}
	ip := net.ParseIP("1.2.3.4")
	result := RunSetUp("age2", ip, true, cert, "/logs", nil, nil)
	if result == nil {
		t.Fatal("nil result")
	}
	// Check captured args contain game and ip
	foundGame := false
	for _, a := range captured.Args {
		if a == "--game=age2" {
			foundGame = true
		}
	}
	if !foundGame {
		t.Errorf("expected --game=age2 in args, got %v", captured.Args)
	}
	// AsAdmin should be true
	if !captured.AsAdmin {
		t.Error("AsAdmin should be true")
	}
	// File should be set
	if captured.File == "" {
		t.Error("File empty")
	}
}

func TestRunSetUp_NilCert(t *testing.T) {
	orig := execFn
	execFn = func(o exec.Options) *exec.Result { return &exec.Result{} }
	defer func() { execFn = orig }()
	ip := net.ParseIP("10.0.0.1")
	res := RunSetUp("age2", ip, false, nil, "", nil, nil)
	if res == nil {
		t.Fatal("nil")
	}
}

func TestRunRevert_FailFastTrue(t *testing.T) {
	orig := execFn
	var captured exec.Options
	execFn = func(o exec.Options) *exec.Result {
		captured = o
		return &exec.Result{}
	}
	defer func() { execFn = orig }()
	RunRevert(true, true, true, "/logs", nil, nil)
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
	orig := execFn
	var captured exec.Options
	execFn = func(o exec.Options) *exec.Result {
		captured = o
		return &exec.Result{}
	}
	defer func() { execFn = orig }()
	RunRevert(false, false, false, "", nil, nil)
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
	orig := execFn
	execFn = func(o exec.Options) *exec.Result { return &exec.Result{Pid: 123} }
	defer func() { execFn = orig }()

	file, result := RunFlushCacheAgent(true, false, "/logs", nil, nil)
	if file == "" {
		t.Error("file empty")
	}
	if result.Pid != 123 {
		t.Errorf("Pid=%d want 123", result.Pid)
	}
	if !result.Success() {
		// Pid 123 with no Err and not Wait => Success checks Pid !=0
	}
	file2, result2 := RunFlushCache(false, true, "/logs", nil, nil)
	if file2 == "" {
		t.Error("file2 empty")
	}
	// RunFlushCache uses Wait true, so without Err it will be success only if ExitCode ==0
	// Our mock returns Pid 123 but Wait true requires ExitCode check; however mock returns ExitCode 0 by default
	if result2 == nil {
		t.Fatal("nil")
	}
}

func TestRun_OptionsFnMutation(t *testing.T) {
	orig := execFn
	execFn = func(o exec.Options) *exec.Result {
		if o.File != "mutated" {
			t.Errorf("optionsFn not applied, File=%q", o.File)
		}
		return &exec.Result{}
	}
	defer func() { execFn = orig }()
	RunSetUp("age2", nil, false, nil, "", nil, func(o *exec.Options) { o.File = "mutated" })
}

func TestRun_OutRedirectionNonWindows(t *testing.T) {
	orig := execFn
	var captured exec.Options
	execFn = func(o exec.Options) *exec.Result {
		captured = o
		return &exec.Result{}
	}
	defer func() { execFn = orig }()
	origAdmin := isAdminFn
	isAdminFn = func() bool { return true }
	defer func() { isAdminFn = origAdmin }()

	var buf io.Writer = &testWriter{}
	RunRevert(true, false, true, "", buf, nil)
	if captured.Stdout != buf {
		t.Error("Stdout should be set when isAdmin true")
	}
}

type testWriter struct{}

func (t *testWriter) Write(p []byte) (n int, err error) { return len(p), nil }

func TestRunFlushCache_OptionsFn(t *testing.T) {
	orig := execFn
	execFn = func(o exec.Options) *exec.Result {
		if o.File != "custom" {
			t.Errorf("optionsFn File not mutated")
		}
		return &exec.Result{}
	}
	defer func() { execFn = orig }()
	RunFlushCache(true, false, "", nil, func(o *exec.Options) { o.File = "custom" })
}
