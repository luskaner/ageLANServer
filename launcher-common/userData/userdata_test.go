package userData

import (
	"path/filepath"
	"testing"

	"github.com/luskaner/ageLANServer/common/game"
)

func TestNewPathJoinsPerGame(t *testing.T) {
	cases := []struct {
		gameId string
		sub    string
		prefix string
	}{
		{game.AoE1, "Age of Empires DE", "Games"},
		{game.AoE2, "Age of Empires 2 DE", "Games"},
		{game.AoE3, "Age of Empires 3 DE", "Games"},
		{game.AoE4, "Age of Empires IV", "My Games"},
		{game.AoM, "Age of Mythology Retold", "Games"},
	}
	for _, tc := range cases {
		p := NewPath("base", tc.gameId)
		want := filepath.Join("base", tc.prefix, tc.sub)
		if p.String() != want {
			t.Errorf("%s: path = %q, want %q", tc.gameId, p.String(), want)
		}
		if p.GameId() != tc.gameId {
			t.Errorf("%s: gameId mismatch", tc.gameId)
		}
	}
}

func TestTypFromExtension(t *testing.T) {
	for _, tc := range []struct {
		path string
		want int
	}{
		{"profile", TypeActive},
		{"profile.lan", TypeServer},
		{"profile.bak", TypeBackup},
		{filepath.Join("dir", "p.bak"), TypeBackup},
	} {
		if got, _ := typ(tc.path); got != tc.want {
			t.Errorf("typ(%q) = %d, want %d", tc.path, got, tc.want)
		}
	}
}

func TestSuffixRoundTrip(t *testing.T) {
	if suffix(TypeActive) != "" {
		t.Fatal("active must have empty suffix")
	}
	if suffix(TypeServer) != serverSuffix || suffix(TypeBackup) != backupSuffix {
		t.Fatal("server/backup suffixes mismatched")
	}
}

func TestTransformPathMatrix(t *testing.T) {
	active := filepath.Join("u", "prof")
	server := active + serverSuffix
	backup := active + backupSuffix

	// Identity transitions.
	if ok, p := TransformPath(active, TypeActive, TypeActive); !ok || p != active {
		t.Errorf("active->active: %v %q", ok, p)
	}
	// Forward transitions.
	if ok, p := TransformPath(active, TypeActive, TypeServer); !ok || p != server {
		t.Errorf("active->server: %v %q", ok, p)
	}
	if ok, p := TransformPath(server, TypeServer, TypeBackup); !ok || p != backup {
		t.Errorf("server->backup: %v %q, want %q", ok, p, backup)
	}
	if ok, p := TransformPath(backup, TypeBackup, TypeActive); !ok || p != active {
		t.Errorf("backup->active: %v %q, want %q", ok, p, active)
	}
	// Mismatched source type fails.
	if ok, _ := TransformPath(active, TypeBackup, TypeServer); ok {
		t.Error("active path must not transform from backup source type")
	}
}
