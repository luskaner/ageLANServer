package models

import (
	"strings"
	"testing"
)

func TestStaticBattleServersPersistAndLoad(t *testing.T) {
	dir := t.TempDir()
	store, err := InitializeStaticBattleServerStore(dir)
	if err != nil {
		t.Fatal(err)
	}
	record := StaticBattleServer{
		GameID:        "age2",
		Region:        "remote-east",
		Name:          "Remote East",
		IPv4:          "192.168.10.25",
		BsPort:        30301,
		WebSocketPort: 30302,
		OutOfBandPort: 30303,
	}
	if err = SaveStaticBattleServer(record); err != nil {
		t.Fatal(err)
	}
	loaded, err := StaticBattleServers("age2")
	if err != nil {
		t.Fatal(err)
	}
	if len(loaded) != 1 || loaded[0] != record {
		t.Fatalf("stored records = %#v, want %#v", loaded, record)
	}
	gameServers, err := LoadStaticBattleServers("age2")
	if err != nil {
		t.Fatal(err)
	}
	if len(gameServers) != 1 || gameServers[0].Region() != record.Region {
		t.Fatalf("loaded game servers = %#v", gameServers)
	}
	if description := gameServers[0].String(); !strings.Contains(description, record.IPv4) {
		t.Fatalf("server description %q does not include its fixed IP", description)
	}
	if err = store.Close(); err != nil {
		t.Fatal(err)
	}

	store, err = InitializeStaticBattleServerStore(dir)
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { _ = store.Close() })
	persisted, err := StaticBattleServers("age2")
	if err != nil {
		t.Fatal(err)
	}
	if len(persisted) != 1 || persisted[0] != record {
		t.Fatalf("reopened records = %#v, want %#v", persisted, record)
	}
}

func TestInitializeBattleServersIncludesPersistentServers(t *testing.T) {
	dir := t.TempDir()
	store, err := InitializeStaticBattleServerStore(dir)
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { _ = store.Close() })
	region := "static-test-startup"
	if err = SaveStaticBattleServer(StaticBattleServer{
		GameID: "age1", Region: region, IPv4: "192.168.10.25", BsPort: 30401, WebSocketPort: 30402,
	}); err != nil {
		t.Fatal(err)
	}
	previous, existed := BattleServersStore["age1"]
	t.Cleanup(func() {
		if existed {
			BattleServersStore["age1"] = previous
		} else {
			delete(BattleServersStore, "age1")
		}
	})
	if err = InitializeBattleServers("age1", nil); err != nil {
		t.Fatal(err)
	}
	found := false
	for _, server := range BattleServersStore["age1"] {
		if server.Region() == region {
			found = true
			break
		}
	}
	if !found {
		t.Fatalf("persistent battle server %q was not loaded at startup", region)
	}
}

func TestSaveStaticBattleServerValidatesPortsAndAddress(t *testing.T) {
	store, err := InitializeStaticBattleServerStore(t.TempDir())
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { _ = store.Close() })

	tests := []struct {
		name   string
		server StaticBattleServer
	}{
		{
			name: "invalid address",
			server: StaticBattleServer{
				GameID: "age2", Region: "bad-ip", IPv4: "not-an-ip", BsPort: 1, WebSocketPort: 2,
			},
		},
		{
			name: "duplicate ports",
			server: StaticBattleServer{
				GameID: "age2", Region: "same-port", IPv4: "192.168.10.25", BsPort: 30301, WebSocketPort: 30301,
			},
		},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			if err := SaveStaticBattleServer(test.server); err == nil {
				t.Fatal("expected invalid static battle server to be rejected")
			}
		})
	}
}
