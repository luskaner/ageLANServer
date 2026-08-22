package cmdUtils

import (
	"net"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/luskaner/ageLANServer/common/battleServer"
)

func TestGeneratePortsAsNeededPreservesExplicit(t *testing.T) {
	gen, err := GeneratePortsAsNeeded([]int{30001, 0, 30003})
	if err != nil {
		t.Fatal(err)
	}
	if gen[0] != 30001 || gen[2] != 30003 {
		t.Fatalf("explicit ports altered: %v", gen)
	}
	if gen[1] <= 0 || gen[1] == 30001 || gen[1] == 30003 {
		t.Fatalf("generated port invalid or collides: %v", gen)
	}
}

func TestGeneratePortsAsNeededAllGeneratedDistinct(t *testing.T) {
	gen, err := GeneratePortsAsNeeded([]int{0, 0, 0})
	if err != nil {
		t.Fatal(err)
	}
	if len(gen) != 3 {
		t.Fatalf("len = %d", len(gen))
	}
	seen := map[int]bool{}
	for _, p := range gen {
		if p <= 0 {
			t.Fatalf("non-positive generated port %d", p)
		}
		if seen[p] {
			t.Fatalf("duplicated generated port %d in %v", p, gen)
		}
		seen[p] = true
	}
}

func TestAvailableFalseWhileBound(t *testing.T) {
	listener, err := net.ListenTCP("tcp4", &net.TCPAddr{IP: net.IPv4zero})
	if err != nil {
		t.Skip("cannot bind ephemeral listener")
	}
	defer func() { _ = listener.Close() }()
	port := listener.Addr().(*net.TCPAddr).Port

	if Available(port) {
		t.Fatalf("port %d is bound; Available must be false", port)
	}

	// After releasing it, it becomes available again (best effort: the OS may
	// hand the port to someone else, so only assert on the negative case).
	_ = listener.Close()
}

// liveValidateConfig builds a config that passes Validate: live PID and two
// ports actually bound by the test (Validate probes them).
func liveValidateConfig(t *testing.T) (battleServer.Config, func()) {
	t.Helper()
	ln1, err := net.ListenTCP("tcp4", &net.TCPAddr{IP: net.IPv4zero})
	if err != nil {
		t.Skip("cannot bind ephemeral listeners")
	}
	ln2, err := net.ListenTCP("tcp4", &net.TCPAddr{IP: net.IPv4zero})
	if err != nil {
		_ = ln1.Close()
		t.Skip("cannot bind ephemeral listeners")
	}
	cfg := battleServer.Config{
		Base: battleServer.Base{
			Region:        "Mi-Región",
			Name:          "My Server",
			IPv4:          "127.0.0.1",
			BsPort:        ln1.Addr().(*net.TCPAddr).Port,
			WebSocketPort: ln2.Addr().(*net.TCPAddr).Port,
		},
		PID: uint32(os.Getpid()),
	}
	return cfg, func() {
		_ = ln1.Close()
		_ = ln2.Close()
	}
}

func TestWriteConfigAndExistingServersRoundTrip(t *testing.T) {
	cfg, closePorts := liveValidateConfig(t)
	defer closePorts()

	gameId := "unit-writer"
	dir := battleServer.Folder(gameId)
	_ = os.RemoveAll(dir) // start clean regardless of leftovers from crashed runs
	if err := os.MkdirAll(dir, 0755); err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { _ = os.RemoveAll(dir) })

	if err := WriteConfig(gameId, cfg); err != nil {
		t.Fatal(err)
	}

	err, names, regions := ExistingServers(gameId)
	if err != nil {
		t.Fatal(err)
	}
	if !names.Contains(strings.ToLower(cfg.Name)) {
		t.Fatalf("names = %v missing %q", names, strings.ToLower(cfg.Name))
	}
	if !regions.Contains(strings.ToLower(cfg.Region)) {
		t.Fatalf("regions = %v missing %q", regions, strings.ToLower(cfg.Region))
	}
}

func TestRemoveKillsNothingAndDeletesFile(t *testing.T) {
	// Dead PID: Kill must treat it as already-dead and return true without
	// signalling anything (a live self-PID here would kill the test binary).
	cfg := battleServer.Config{
		Base: battleServer.Base{Region: "r", Name: "n", IPv4: "1.2.3.4", BsPort: 1, WebSocketPort: 2},
		PID:  4_000_000_000,
	}

	gameId := "unit-remover"
	dir := battleServer.Folder(gameId)
	_ = os.RemoveAll(dir) // start clean regardless of leftovers from crashed runs
	if err := os.MkdirAll(dir, 0755); err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { _ = os.RemoveAll(dir) })

	if err := WriteConfig(gameId, cfg); err != nil {
		t.Fatal(err)
	}

	configs, confErr := battleServer.Configs(gameId, false, false)
	if confErr != nil || len(configs) != 1 {
		t.Fatalf("configs = %d err = %v", len(configs), confErr)
	}
	if !Remove(gameId, configs, false) {
		t.Fatal("Remove reported nothing removed")
	}
	if _, statErr := os.Stat(filepath.Join(dir, battleServer.Name(1))); !os.IsNotExist(statErr) {
		t.Fatal("config file not removed")
	}
}
