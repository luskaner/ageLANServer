//go:build windows

package game

import (
	"testing"
)

func TestSupportedGames(t *testing.T) {
	expected := []string{AoE1, AoE2, AoE3, AoE4, AoM}
	if SupportedGames.Cardinality() != len(expected) {
		t.Errorf("SupportedGames.Cardinality() = %d, want %d", SupportedGames.Cardinality(), len(expected))
	}
	for _, g := range expected {
		if !SupportedGames.Contains(g) {
			t.Errorf("SupportedGames does not contain %q", g)
		}
	}
}

func TestAllGamesEqualsSupportedGames(t *testing.T) {
	if AllGames.Cardinality() != SupportedGames.Cardinality() {
		t.Errorf("AllGames and SupportedGames have different sizes: %d vs %d",
			AllGames.Cardinality(), SupportedGames.Cardinality())
	}
}

func TestGameConstants(t *testing.T) {
	if AoE1 != "age1" {
		t.Errorf("AoE1 = %q, want %q", AoE1, "age1")
	}
	if AoE2 != "age2" {
		t.Errorf("AoE2 = %q, want %q", AoE2, "age2")
	}
	if AoE3 != "age3" {
		t.Errorf("AoE3 = %q, want %q", AoE3, "age3")
	}
	if AoE4 != "age4" {
		t.Errorf("AoE4 = %q, want %q", AoE4, "age4")
	}
	if AoM != "athens" {
		t.Errorf("AoM = %q, want %q", AoM, "athens")
	}
}
