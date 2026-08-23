package executor

import (
	"errors"
	"testing"

	"github.com/luskaner/ageLANServer/common/executor/exec"
)

func TestNewConfigFlushCacheOptionsBothFalse(t *testing.T) {
	if opts := NewConfigFlushCacheOptions(false, "", false, false); opts != nil {
		t.Fatal("expected nil when both ips and certs are false")
	}
}

func TestNewConfigFlushCacheOptionsIPsOnly(t *testing.T) {
	opts := NewConfigFlushCacheOptions(true, "local", false, true)
	if opts == nil {
		t.Fatal("expected non-nil for IPs enabled")
	}
	if !opts.IPs {
		t.Fatal("IPs should be true")
	}
	if opts.Certs {
		t.Fatal("Certs should be false")
	}
}

func TestNewConfigFlushCacheOptionsCertOnlyLinux(t *testing.T) {
	if !isLinux() {
		t.Skip("cert flush is linux-only")
	}
	opts := NewConfigFlushCacheOptions(false, "local", true, false)
	if opts == nil {
		t.Fatal("expected non-nil for certs enabled on linux")
	}
	if !opts.Certs {
		t.Fatal("Certs should be true")
	}
}

func TestNewConfigFlushCacheOptionsCustomHostFileDisablesIPs(t *testing.T) {
	if opts := NewConfigFlushCacheOptions(true, "local", true, false); opts != nil && opts.IPs {
		t.Fatal("custom host file must disable IP flush")
	}
}

func isAdminError(err error) bool { return errors.Is(err, errTest) }

var errTest = errors.New("test")

func isAdminErrorHelper(result *exec.Result) bool {
	return result != nil && result.Err != nil && errors.Is(result.Err, errTest)
}

func isLinux() bool { return false }
