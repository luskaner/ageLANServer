package game

import (
	"testing"

	commonGame "github.com/luskaner/ageLANServer/common/game"
)

func TestSteamProcess(t *testing.T) {
	tests := []struct {
		gameId string
		want   string
	}{
		{commonGame.AoE1, "AoEDE_s.exe"},
		{commonGame.AoE2, "AoE2DE_s.exe"},
		{commonGame.AoE3, "AoE3DE_s.exe"},
		{commonGame.AoE4, "RelicCardinal.exe"},
		{commonGame.AoM, "AoMRT_s.exe"},
		{"unknown", ""},
	}
	for _, tt := range tests {
		t.Run(tt.gameId, func(t *testing.T) {
			if got := steamProcess(tt.gameId); got != tt.want {
				t.Errorf("steamProcess(%q) = %q, want %q", tt.gameId, got, tt.want)
			}
		})
	}
}

func TestGame_SteamProcess(t *testing.T) {
	tests := []struct {
		process string
		want    string
	}{
		{"AoEDE_s.exe", commonGame.AoE1},
		{"AoE2DE_s.exe", commonGame.AoE2},
		{"AoE3DE_s.exe", commonGame.AoE3},
		{"RelicCardinal.exe", commonGame.AoE4},
		{"AoMRT_s.exe", commonGame.AoM},
		{"unknown.exe", ""},
	}
	for _, tt := range tests {
		t.Run(tt.process, func(t *testing.T) {
			if got := Game(tt.process, false); got != tt.want {
				t.Errorf("Game(%q, false) = %q, want %q", tt.process, got, tt.want)
			}
		})
	}
}

func TestProcesses_Steam(t *testing.T) {
	procs := Processes(commonGame.AoE2, true, false, false)
	if len(procs) != 1 {
		t.Fatalf("expected 1 process, got %d: %v", len(procs), procs)
	}
	if procs[0] != "AoE2DE_s.exe" {
		t.Errorf("expected AoE2DE_s.exe, got %s", procs[0])
	}
}

func TestProcesses_None(t *testing.T) {
	procs := Processes(commonGame.AoE2, false, false, false)
	if len(procs) != 0 {
		t.Errorf("expected 0 processes, got %d: %v", len(procs), procs)
	}
}

func TestGame_UnknownProcess(t *testing.T) {
	if got := Game("nonexistent.exe", false); got != "" {
		t.Errorf("Game(\"nonexistent.exe\", false) = %q, want \"\"", got)
	}
}
