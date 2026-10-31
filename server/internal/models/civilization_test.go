package models

import "testing"

func TestCivilizationName(t *testing.T) {
	tests := []struct {
		gameID string
		raceID int32
		want   string
	}{
		{gameID: "age2", raceID: 0, want: "Aztecs"},
		{gameID: "age2", raceID: 2, want: "Britons"},
		{gameID: "age2", raceID: 26, want: "Portuguese"},
		{gameID: "age2", raceID: 44, want: "Georgians"},
		{gameID: "age2", raceID: 45, want: "Civilization 45"},
		{gameID: "age2", raceID: -1, want: "Unknown"},
		{gameID: "age4", raceID: 1, want: "Civilization 1"},
	}
	for _, test := range tests {
		if got := CivilizationName(test.gameID, test.raceID); got != test.want {
			t.Errorf("CivilizationName(%q, %d) = %q, want %q", test.gameID, test.raceID, got, test.want)
		}
	}
}

func TestCivilizationIconPath(t *testing.T) {
	tests := []struct {
		gameID string
		raceID int32
		want   string
	}{
		{gameID: "age2", raceID: 0, want: "/static/images/civs/aztecs.png"},
		{gameID: "age2", raceID: 44, want: "/static/images/civs/georgians.png"},
		{gameID: "age2", raceID: 45, want: "/static/images/civs/unknown.svg"},
		{gameID: "age4", raceID: 1, want: "/static/images/civs/unknown.svg"},
	}
	for _, test := range tests {
		if got := CivilizationIconPath(test.gameID, test.raceID); got != test.want {
			t.Errorf("CivilizationIconPath(%q, %d) = %q, want %q", test.gameID, test.raceID, got, test.want)
		}
	}
}
