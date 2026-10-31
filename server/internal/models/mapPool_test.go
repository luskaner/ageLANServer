package models

import (
	"encoding/json"
	"os"
	"path/filepath"
	"testing"
)

func TestMapPoolPersistsForMatchmakingAndNormalizesIdentity(t *testing.T) {
	dir := t.TempDir()
	store := NewMapPoolStore(dir)
	store.gameID = "age2"
	pool := MapPool{
		GameID:  "age2",
		QueueID: AutomatchQueueID(),
		Mode:    MapPoolModeRotation,
		Maps: []MapPoolEntry{
			{MapID: 9, Name: "Arabia", Weight: 1, Enabled: true},
			{MapID: 10, Name: "Arena", Weight: 2, Enabled: true},
		},
	}
	if err := store.Set(pool); err != nil {
		t.Fatal(err)
	}
	persisted, ok := store.Get(AutomatchQueueID())
	if !ok || persisted.GameID != "age2" || persisted.Mode != MapPoolModeRotation {
		t.Fatalf("stored pool = %#v, ok=%t", persisted, ok)
	}
	selected, ok := store.PickSelection(AutomatchQueueID(), []AvailableMap{
		{ID: 9, Name: "Arabia", MaxPlayers: 2},
		{ID: 10, Name: "Arena", MaxPlayers: 2},
	})
	if !ok || selected.ID != 9 {
		t.Fatalf("first rotation map = %#v, ok=%t", selected, ok)
	}
	if err := store.Close(); err != nil {
		t.Fatal(err)
	}

	reopened := NewMapPoolStore(dir)
	reopened.gameID = "age2"
	t.Cleanup(func() { _ = reopened.Close() })
	persisted, ok = reopened.Get(AutomatchQueueID())
	if !ok || persisted.Cursor != 1 {
		t.Fatalf("reopened pool cursor = %#v, ok=%t", persisted, ok)
	}
}

func TestMapPoolRejectsExcessiveWeight(t *testing.T) {
	pool := MapPool{
		Mode: MapPoolModeWeighted,
		Maps: []MapPoolEntry{
			{MapID: 1, Name: "Arabia", Weight: maxMapPoolWeight + 1, Enabled: true},
		},
	}
	if err := pool.Normalize(); err == nil {
		t.Fatal("expected an out-of-range map weight to be rejected")
	}
}

func TestDisplayMapName(t *testing.T) {
	tests := map[string]string{
		"goldenpit.rms2":   "Golden Pit",
		"FourLakes.rms":    "Four Lakes",
		"Black_Forest.rms": "Black Forest",
	}
	for input, want := range tests {
		if got := DisplayMapName(input); got != want {
			t.Errorf("DisplayMapName(%q) = %q, want %q", input, got, want)
		}
	}
}

func TestMapIconPath(t *testing.T) {
	tests := map[string]string{
		"Arabia.rms":     "/static/images/maps/arabia.png",
		"GoldenPit.rms2": "/static/images/maps/golden-pit.png",
		"UnknownMap.rms": "/static/images/maps/unknown-map.png",
		"   ":            "/static/images/map-placeholder.svg",
	}
	for input, want := range tests {
		if got := MapIconPath(input); got != want {
			t.Errorf("MapIconPath(%q) = %q, want %q", input, got, want)
		}
	}
}

func TestAutomatchMapsForPlayersIncludesCatalogMapsAndPrefersQueueEntry(t *testing.T) {
	payload := []any{0, []any{
		[]any{1, "Arabia.rms", 2, 2, 6},
		[]any{69, "Arabia.rms", 2, 2, 2},
		[]any{70, "Ghost_Lake.rms", 2, 2, 2},
		[]any{71, "TeamMap.rms", 2, 4, 2},
	}}

	maps := AutomatchMapsForPlayers(payload, 6, 2)
	if len(maps) != 2 {
		t.Fatalf("available maps = %+v, want the two unique two-player maps", maps)
	}
	if maps[0].Name != "Arabia.rms" || maps[1].Name != "Ghost_Lake.rms" {
		t.Fatalf("available map names = %+v", maps)
	}
	if maps[0].ID != 1 || maps[0].LeaderboardID != 6 {
		t.Fatalf("queue-specific representative = %+v, want Arabia's queue 6 entry", maps[0])
	}
	if maps[0].Icon != "/static/images/maps/arabia.png" {
		t.Errorf("Arabia icon = %q, want its dashboard thumbnail URL", maps[0].Icon)
	}
}

func TestFilterAutomatchMapsPublishesSelectedCatalogMapsAsQueueMaps(t *testing.T) {
	payload := []any{0, []any{
		[]any{1, "Arabia.rms", 2, 2, 6},
		[]any{3, "Arena.rms", 2, 2, 6},
		[]any{69, "Ghost_Lake.rms", 2, 2, 2},
		[]any{80, "TeamMap.rms", 2, 4, 6},
	}}
	filtered := FilterAutomatchMaps(payload, 6,
		map[int32]struct{}{69: {}},
		map[string]struct{}{"ghost lake": {}},
	)
	entries := filtered[1].([]any)
	foundGhostLake, foundArena, foundTeamMap := false, false, false
	for _, raw := range entries {
		entry := raw.([]any)
		name := entry[1].(string)
		leaderboard := int32(entry[4].(int))
		switch name {
		case "Ghost_Lake.rms":
			if leaderboard == 6 {
				foundGhostLake = true
			}
		case "Arena.rms":
			foundArena = true
		case "TeamMap.rms":
			foundTeamMap = true
		}
	}
	if !foundGhostLake || foundArena || !foundTeamMap {
		t.Fatalf("filtered map entries = %#v; want selected catalog map, no unselected 1v1 map, and other queue entries unchanged", entries)
	}
}

func TestAge2MapPoolListsEveryUniqueTwoPlayerMap(t *testing.T) {
	raw, err := os.ReadFile(filepath.Join("..", "..", "resources", "responses", "age2", "automatchMaps.json"))
	if err != nil {
		t.Fatal(err)
	}
	var payload []any
	if err = json.Unmarshal(raw, &payload); err != nil {
		t.Fatal(err)
	}
	maps := AutomatchMapsForPlayers(payload, AutomatchQueueLeaderboardID(), 2)
	if len(maps) != 45 {
		t.Fatalf("1v1 map-pool options = %d, want all 45 unique two-player maps from the game catalog", len(maps))
	}
	filtered := FilterAutomatchMaps(payload, AutomatchQueueLeaderboardID(),
		map[int32]struct{}{70: {}},
		map[string]struct{}{"ghost lake": {}},
	)
	foundSelected1v1Map := false
	for _, rawEntry := range filtered[1].([]any) {
		entry := rawEntry.([]any)
		id, _ := numberToInt32(entry[0])
		queue, _ := numberToInt32(entry[4])
		if id == 70 && entry[1] == "Ghost_Lake.rms" && queue == AutomatchQueueLeaderboardID() {
			foundSelected1v1Map = true
			break
		}
	}
	if !foundSelected1v1Map {
		t.Fatal("selected map from another queue was not published as a 1v1 map")
	}
}
