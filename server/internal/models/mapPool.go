package models

import (
	"database/sql"
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"sort"
	"strings"
	"sync"
	"time"
	"unicode"

	"github.com/luskaner/ageServer/server/internal"
	"github.com/luskaner/ageServer/server/internal/events"
	"github.com/luskaner/ageServer/server/internal/logger"
)

// Map pool modes.
const (
	maxMapPoolWeight      = 1000
	maxMapPoolTotalWeight = 100000
	// MapPoolModeRandom picks a uniformly random enabled map.
	MapPoolModeRandom = "random"
	// MapPoolModeRotation walks the enabled maps in order, one per match, and
	// wraps around at the end.
	MapPoolModeRotation = "rotation"
	// MapPoolModeWeighted picks a random enabled map using its weight.
	MapPoolModeWeighted = "weighted"
)

// MapPoolEntry is one map in an admin-configured pool.
type MapPoolEntry struct {
	MapID   int32  `json:"id"`
	Name    string `json:"name"`
	Weight  int    `json:"weight"`
	Enabled bool   `json:"enabled"`
}

// MapPool is the configured map pool for one automatch queue.
type MapPool struct {
	GameID    string         `json:"game"`
	QueueID   int32          `json:"queue"`
	Mode      string         `json:"mode"`
	Maps      []MapPoolEntry `json:"maps"`
	Cursor    int            `json:"cursor"`
	UpdatedAt time.Time      `json:"updatedAt"`
}

// EnabledMaps returns the enabled entries in their configured order.
func (p *MapPool) EnabledMaps() []MapPoolEntry {
	if p == nil {
		return nil
	}
	enabled := make([]MapPoolEntry, 0, len(p.Maps))
	for _, entry := range p.Maps {
		if entry.Enabled && entry.Name != "" {
			enabled = append(enabled, entry)
		}
	}
	return enabled
}

// Normalize validates a pool and fills in defaults. It is the single place that
// decides what a valid pool looks like, so the admin API, the HTTP decoder and
// the matchmaking path all agree.
func (p *MapPool) Normalize() error {
	if p == nil {
		return errors.New("map pool is nil")
	}
	p.Mode = strings.ToLower(strings.TrimSpace(p.Mode))
	switch p.Mode {
	case "":
		p.Mode = MapPoolModeRandom
	case MapPoolModeRandom, MapPoolModeRotation, MapPoolModeWeighted:
	default:
		return fmt.Errorf("unknown map pool mode %q", p.Mode)
	}

	seen := make(map[int32]struct{}, len(p.Maps))
	cleaned := make([]MapPoolEntry, 0, len(p.Maps))
	totalWeight := 0
	for index, entry := range p.Maps {
		entry.Name = strings.TrimSpace(entry.Name)
		if entry.Name == "" {
			return fmt.Errorf("map %d has no name", index)
		}
		if entry.MapID == 0 {
			return fmt.Errorf("map %q has no id", entry.Name)
		}
		if _, duplicate := seen[entry.MapID]; duplicate {
			return fmt.Errorf("map id %d is duplicated", entry.MapID)
		}
		seen[entry.MapID] = struct{}{}
		if entry.Weight <= 0 {
			entry.Weight = 1
		}
		if entry.Weight > maxMapPoolWeight {
			return fmt.Errorf("map %q weight must not exceed %d", entry.Name, maxMapPoolWeight)
		}
		totalWeight += entry.Weight
		if totalWeight > maxMapPoolTotalWeight {
			return fmt.Errorf("total map weight must not exceed %d", maxMapPoolTotalWeight)
		}
		cleaned = append(cleaned, entry)
	}
	p.Maps = cleaned
	if len(p.EnabledMaps()) == 0 {
		return errors.New("the map pool has no enabled maps")
	}
	if p.Cursor < 0 {
		p.Cursor = 0
	}
	if p.UpdatedAt.IsZero() {
		p.UpdatedAt = time.Now().UTC()
	}
	return nil
}

// MapPoolStore persists map pools in the same SQLite database the ladder uses
// (resources/userData/<game>/automatch2.db), so the web dashboard and the
// matchmaking code read and write a single copy of the truth.
type MapPoolStore struct {
	mu     sync.RWMutex
	pickMu sync.Mutex
	db     *sql.DB
	path   string
	gameID string
}

// NewMapPoolStore opens (and creates when missing) the pool database in dataDir.
func NewMapPoolStore(dataDir string) *MapPoolStore {
	s := &MapPoolStore{path: filepath.Join(dataDir, "automatch2.db")}
	if err := os.MkdirAll(dataDir, 0o755); err != nil {
		logger.Printf("mapPool: cannot create data dir %s: %s\n", dataDir, err)
		return s
	}
	db, err := sql.Open("sqlite", s.path)
	if err != nil {
		logger.Printf("mapPool: cannot open %s: %s\n", s.path, err)
		return s
	}
	// Keep a single connection: the pool code does very little work and this
	// avoids SQLITE_BUSY against the ladder's own connection.
	db.SetMaxOpenConns(1)
	for _, pragma := range []string{
		"PRAGMA journal_mode=WAL",
		"PRAGMA synchronous=NORMAL",
		"PRAGMA busy_timeout=5000",
	} {
		if _, err = db.Exec(pragma); err != nil {
			logger.Printf("mapPool: %s failed: %s\n", pragma, err)
		}
	}
	if _, err = db.Exec(`CREATE TABLE IF NOT EXISTS map_pools (
		queue_id   INTEGER PRIMARY KEY,
		mode       TEXT    NOT NULL DEFAULT 'random',
		entries    TEXT    NOT NULL DEFAULT '[]',
		cursor     INTEGER NOT NULL DEFAULT 0,
		updated_at INTEGER NOT NULL DEFAULT 0
	)`); err != nil {
		logger.Printf("mapPool: schema migration failed: %s\n", err)
		_ = db.Close()
		return s
	}
	s.db = db
	return s
}

// NewMapPoolStoreForGame keeps the pool next to the rest of the per-game user
// data, which is where the ladder database already lives.
func NewMapPoolStoreForGame(gameId string) *MapPoolStore {
	store := NewMapPoolStore(filepath.Join(userDataFolder, gameId))
	store.gameID = gameId
	return store
}

// Path returns the database file backing the store.
func (s *MapPoolStore) Path() string { return s.path }

// Close closes the underlying database.
func (s *MapPoolStore) Close() error {
	s.mu.Lock()
	defer s.mu.Unlock()
	if s.db == nil {
		return nil
	}
	err := s.db.Close()
	s.db = nil
	return err
}

// Get returns the stored pool for a queue.
func (s *MapPoolStore) Get(queueID int32) (*MapPool, bool) {
	if s == nil {
		return nil, false
	}
	s.mu.RLock()
	defer s.mu.RUnlock()
	if s.db == nil {
		return nil, false
	}
	var (
		mode      string
		raw       string
		cursor    int
		updatedAt int64
	)
	err := s.db.QueryRow(
		`SELECT mode, entries, cursor, updated_at FROM map_pools WHERE queue_id = ?`,
		queueID,
	).Scan(&mode, &raw, &cursor, &updatedAt)
	if err != nil {
		if !errors.Is(err, sql.ErrNoRows) {
			logger.Printf("mapPool: read queue %d: %s\n", queueID, err)
		}
		return nil, false
	}
	pool := &MapPool{
		GameID:  s.gameID,
		QueueID: queueID,
		Mode:    mode,
		Cursor:  cursor,
	}
	if err = json.Unmarshal([]byte(raw), &pool.Maps); err != nil {
		logger.Printf("mapPool: decode queue %d: %s\n", queueID, err)
		return nil, false
	}
	if updatedAt > 0 {
		pool.UpdatedAt = time.Unix(updatedAt, 0).UTC()
	}
	return pool, true
}

// Set validates and stores a pool.
func (s *MapPoolStore) Set(pool MapPool) error {
	if s == nil {
		return errors.New("no map pool store")
	}
	if err := pool.Normalize(); err != nil {
		return err
	}
	raw, err := json.Marshal(pool.Maps)
	if err != nil {
		return err
	}
	s.mu.Lock()
	defer s.mu.Unlock()
	if s.db == nil {
		return errors.New("map pool database is not open")
	}
	_, err = s.db.Exec(
		`INSERT INTO map_pools (queue_id, mode, entries, cursor, updated_at)
		 VALUES (?, ?, ?, ?, ?)
		 ON CONFLICT(queue_id) DO UPDATE SET
		   mode = excluded.mode,
		   entries = excluded.entries,
		   cursor = excluded.cursor,
		   updated_at = excluded.updated_at`,
		pool.QueueID, pool.Mode, string(raw), pool.Cursor, pool.UpdatedAt.Unix(),
	)
	if err != nil {
		return err
	}
	events.Publish("pool.updated", pool.GameID, pool)
	return nil
}

// Reset removes the pool of a queue, restoring the built-in behaviour (every
// map of automatchMaps.json that the players did not veto).
func (s *MapPoolStore) Reset(queueID int32) error {
	if s == nil {
		return errors.New("no map pool store")
	}
	s.mu.Lock()
	defer s.mu.Unlock()
	if s.db == nil {
		return errors.New("map pool database is not open")
	}
	if _, err := s.db.Exec(`DELETE FROM map_pools WHERE queue_id = ?`, queueID); err != nil {
		return err
	}
	events.Publish("pool.reset", "", map[string]any{"queue": queueID})
	return nil
}

// SetCursor stores the rotation cursor without touching the rest of the pool.
func (s *MapPoolStore) SetCursor(queueID int32, cursor int) error {
	pool, ok := s.Get(queueID)
	if !ok {
		return fmt.Errorf("no pool configured for queue %d", queueID)
	}
	if cursor < 0 {
		cursor = 0
	}
	enabled := len(pool.EnabledMaps())
	if enabled > 0 {
		cursor %= enabled
	}
	pool.Cursor = cursor
	return s.Set(*pool)
}

// Pick chooses the map for the next match of a queue. It returns false when no
// pool is configured, so the caller keeps its built-in behaviour.
func (s *MapPoolStore) Pick(queueID int32) (MapPoolEntry, bool) {
	if s == nil {
		return MapPoolEntry{}, false
	}
	s.pickMu.Lock()
	defer s.pickMu.Unlock()
	pool, ok := s.Get(queueID)
	if !ok {
		return MapPoolEntry{}, false
	}
	enabled := pool.EnabledMaps()
	if len(enabled) == 0 {
		return MapPoolEntry{}, false
	}

	var chosen MapPoolEntry
	switch pool.Mode {
	case MapPoolModeRotation:
		index := pool.Cursor % len(enabled)
		if index < 0 {
			index = 0
		}
		chosen = enabled[index]
		next := (index + 1) % len(enabled)
		if err := s.setCursorRaw(queueID, next); err != nil {
			logger.Printf("mapPool: advance rotation queue %d: %s\n", queueID, err)
		}
		events.Publish("pool.rotated", pool.GameID, map[string]any{
			"queue": queueID,
			"map":   chosen.Name,
			"id":    chosen.MapID,
			"next":  next,
		})
	case MapPoolModeWeighted:
		total := 0
		for _, entry := range enabled {
			total += entry.Weight
		}
		internal.WithRng(func(rng *internal.RandReader) {
			roll := rng.IntN(total)
			for _, entry := range enabled {
				if roll < entry.Weight {
					chosen = entry
					return
				}
				roll -= entry.Weight
			}
			chosen = enabled[len(enabled)-1]
		})
	default:
		internal.WithRng(func(rng *internal.RandReader) {
			chosen = enabled[rng.IntN(len(enabled))]
		})
	}
	return chosen, true
}

// PickSelection chooses one of the selectable maps of a queue according to the
// configured pool. selectable is the veto-filtered candidate list, so a pool
// map the players vetoed is never returned.
//
// It returns false when the queue has no pool, keeping the caller's built-in
// behaviour untouched.
func (s *MapPoolStore) PickSelection(queueID int32, selectable []AvailableMap) (AvailableMap, bool) {
	if s == nil || len(selectable) == 0 {
		return AvailableMap{}, false
	}
	s.pickMu.Lock()
	defer s.pickMu.Unlock()
	pool, ok := s.Get(queueID)
	if !ok {
		return AvailableMap{}, false
	}
	enabled := pool.EnabledMaps()
	if len(enabled) == 0 {
		return AvailableMap{}, false
	}

	byID := make(map[int32]AvailableMap, len(selectable))
	byName := make(map[string]AvailableMap, len(selectable))
	for _, m := range selectable {
		byID[m.ID] = m
		byName[normalizeMapName(m.Name)] = m
	}
	find := func(entry MapPoolEntry) (AvailableMap, bool) {
		if m, ok := byID[entry.MapID]; ok {
			return m, true
		}
		if m, ok := byName[normalizeMapName(entry.Name)]; ok {
			return m, true
		}
		return AvailableMap{}, false
	}

	if pool.Mode == MapPoolModeRotation {
		for offset := 0; offset < len(enabled); offset++ {
			index := (pool.Cursor + offset) % len(enabled)
			chosen, ok := find(enabled[index])
			if !ok {
				continue
			}
			next := (index + 1) % len(enabled)
			if err := s.setCursorRaw(queueID, next); err != nil {
				logger.Printf("mapPool: advance rotation queue %d: %s\n", queueID, err)
			}
			events.Publish("pool.rotated", pool.GameID, map[string]any{
				"queue": queueID,
				"map":   chosen.Name,
				"id":    chosen.ID,
				"next":  next,
			})
			return chosen, true
		}
		return AvailableMap{}, false
	}

	// Weighted and random modes only consider pool maps that survived the
	// vetoes, so the weights keep their relative meaning.
	candidates := make([]AvailableMap, 0, len(enabled))
	weights := make([]int, 0, len(enabled))
	for _, entry := range enabled {
		m, ok := find(entry)
		if !ok {
			continue
		}
		weight := entry.Weight
		if weight <= 0 {
			weight = 1
		}
		candidates = append(candidates, m)
		weights = append(weights, weight)
	}
	if len(candidates) == 0 {
		return AvailableMap{}, false
	}

	var chosen AvailableMap
	internal.WithRng(func(rng *internal.RandReader) {
		if pool.Mode != MapPoolModeWeighted {
			chosen = candidates[rng.IntN(len(candidates))]
			return
		}
		total := 0
		for _, weight := range weights {
			total += weight
		}
		roll := rng.IntN(total)
		for index, weight := range weights {
			if roll < weight {
				chosen = candidates[index]
				return
			}
			roll -= weight
		}
		chosen = candidates[len(candidates)-1]
	})
	return chosen, true
}

// FilterAutomatchMaps applies a queue's pool to its map list. For the 1v1
// queue, selected two-player maps published for other queues are exposed as
// entries for this queue as well. Other queues keep their full map lists.
//
// The payload shape is [0, [[id, name, teams, maxPlayers, leaderboardID, ...],
// ...]]. Entries are matched by id first and by normalized name second, so a
// pool stored with a slightly different name spelling still works.
func FilterAutomatchMaps(payload []any, leaderboardID int32, allowedIDs map[int32]struct{}, allowedNames map[string]struct{}) []any {
	if len(payload) < 2 || allowedIDs == nil {
		return payload
	}
	entries, ok := payload[1].([]any)
	if !ok {
		return payload
	}
	filtered := make([]any, 0, len(entries))
	for _, raw := range entries {
		entry, ok := raw.([]any)
		if !ok || len(entry) < 5 {
			filtered = append(filtered, raw)
			continue
		}
		entryLeaderboard, ok := numberToInt32(entry[4])
		if !ok {
			filtered = append(filtered, raw)
			continue
		}
		if entryLeaderboard != leaderboardID {
			// Another queue: the pool does not own it.
			filtered = append(filtered, raw)
			continue
		}
		maxPlayers, maxPlayersOK := numberToInt32(entry[3])
		if !maxPlayersOK || maxPlayers != 2 {
			filtered = append(filtered, raw)
		}
	}
	for _, candidate := range AutomatchMapsForPlayers(payload, leaderboardID, 2) {
		_, allowedByID := allowedIDs[candidate.ID]
		_, allowedByName := allowedNames[normalizeMapName(candidate.Name)]
		if !allowedByID && !allowedByName {
			continue
		}
		for _, raw := range entries {
			entry, ok := raw.([]any)
			if !ok || len(entry) < 5 {
				continue
			}
			id, idOK := numberToInt32(entry[0])
			maxPlayers, maxPlayersOK := numberToInt32(entry[3])
			if !idOK || id != candidate.ID || !maxPlayersOK || maxPlayers != 2 {
				continue
			}
			copyEntry := append([]any(nil), entry...)
			copyEntry[4] = int(leaderboardID)
			filtered = append(filtered, copyEntry)
			break
		}
	}
	result := make([]any, len(payload))
	copy(result, payload)
	result[1] = filtered
	return result
}

// poolMode reports the configured mode of a queue, for logging.
func poolMode(s *MapPoolStore, queueID int32) string {
	if pool, ok := s.Get(queueID); ok {
		return pool.Mode
	}
	return ""
}

// FilterAutomatchMapsByPool applies the pool of a queue to an
// automatchMaps.json payload. It returns the payload untouched when the queue
// has no pool configured.
func FilterAutomatchMapsByPool(payload []any, gameID string, queueID int32) []any {
	if gameID == "" || queueID <= 0 {
		return payload
	}
	ids, names := MapPools(gameID).AllowedMaps(queueID)
	if ids == nil {
		return payload
	}
	return FilterAutomatchMaps(payload, queueID, ids, names)
}

// setCursorRaw persists the rotation cursor without re-publishing the pool.
func (s *MapPoolStore) setCursorRaw(queueID int32, cursor int) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	if s.db == nil {
		return errors.New("map pool database is not open")
	}
	_, err := s.db.Exec(
		`UPDATE map_pools SET cursor = ?, updated_at = ? WHERE queue_id = ?`,
		cursor, time.Now().UTC().Unix(), queueID,
	)
	return err
}

// AllowedMaps returns the ids and names the pool allows for a queue. When no
// pool is configured (or it has no enabled maps) it returns nil, meaning "no
// restriction".
func (s *MapPoolStore) AllowedMaps(queueID int32) (ids map[int32]struct{}, names map[string]struct{}) {
	pool, ok := s.Get(queueID)
	if !ok {
		return nil, nil
	}
	enabled := pool.EnabledMaps()
	if len(enabled) == 0 {
		return nil, nil
	}
	ids = make(map[int32]struct{}, len(enabled))
	names = make(map[string]struct{}, len(enabled))
	for _, entry := range enabled {
		ids[entry.MapID] = struct{}{}
		names[normalizeMapName(entry.Name)] = struct{}{}
	}
	return ids, names
}

// AvailableMap is one selectable map as published by the client resources.
type AvailableMap struct {
	ID            int32  `json:"id"`
	Name          string `json:"name"`
	DisplayName   string `json:"displayName"`
	Icon          string `json:"icon"`
	MaxPlayers    int    `json:"maxPlayers"`
	LeaderboardID int32  `json:"leaderboard"`
}

// AvailableMaps parses automatchMaps.json into the flat list of maps for a
// queue. The client payload shape is
// [0, [[id, name, teams, maxPlayers, leaderboardID, ...], ...]].
func AvailableMaps(resources []any, queueID int32) []AvailableMap {
	_ = queueID
	if len(resources) < 2 {
		return nil
	}
	entries, ok := resources[1].([]any)
	if !ok {
		return nil
	}
	maps := make([]AvailableMap, 0, len(entries))
	for _, raw := range entries {
		entry, ok := raw.([]any)
		if !ok || len(entry) < 5 {
			continue
		}
		id, ok := numberToInt32(entry[0])
		if !ok {
			continue
		}
		name, ok := entry[1].(string)
		if !ok || name == "" {
			continue
		}
		maxPlayers, _ := numberToInt32(entry[3])
		leaderboard, _ := numberToInt32(entry[4])
		maps = append(maps, AvailableMap{
			ID:            id,
			Name:          name,
			DisplayName:   DisplayMapName(name),
			Icon:          MapIconPath(name),
			MaxPlayers:    int(maxPlayers),
			LeaderboardID: leaderboard,
		})
	}
	sort.SliceStable(maps, func(a, b int) bool { return maps[a].ID < maps[b].ID })
	return maps
}

// AutomatchMapsForPlayers returns one representative entry for each map name
// published for the requested player count. Maps already listed for queueID
// take precedence; other queue entries let the admin configure the full game
// catalog for a queue-specific pool.
func AutomatchMapsForPlayers(resources []any, queueID int32, maxPlayers int) []AvailableMap {
	byName := make(map[string]AvailableMap)
	for _, candidate := range AvailableMaps(resources, queueID) {
		if candidate.MaxPlayers != maxPlayers {
			continue
		}
		key := normalizeMapName(candidate.Name)
		current, found := byName[key]
		if !found ||
			(candidate.LeaderboardID == queueID && current.LeaderboardID != queueID) ||
			(candidate.LeaderboardID == current.LeaderboardID && candidate.ID < current.ID) {
			byName[key] = candidate
		}
	}
	maps := make([]AvailableMap, 0, len(byName))
	for _, candidate := range byName {
		maps = append(maps, candidate)
	}
	sort.Slice(maps, func(i, j int) bool {
		left, right := strings.ToLower(maps[i].DisplayName), strings.ToLower(maps[j].DisplayName)
		if left != right {
			return left < right
		}
		return maps[i].ID < maps[j].ID
	})
	return maps
}

func MapIconPath(name string) string {
	displayName := DisplayMapName(name)
	slug := strings.ToLower(mapAssetSlugReplacer.Replace(displayName))
	slug = strings.Trim(slug, "-")
	if slug == "" {
		return "/static/images/map-placeholder.svg"
	}
	return "/static/images/maps/" + slug + ".png"
}

var mapAssetSlugReplacer = strings.NewReplacer(
	" ", "-",
	"_", "-",
	".", "-",
	"/", "-",
	"\\", "-",
)

func DisplayMapName(name string) string {
	for _, extension := range []string{".rms2", ".rms"} {
		if strings.HasSuffix(strings.ToLower(name), extension) {
			name = name[:len(name)-len(extension)]
			break
		}
	}
	switch strings.ToLower(name) {
	case "goldenpit":
		return "Golden Pit"
	case "megarandom":
		return "Mega Random"
	}
	name = mapCamelCaseBoundary.ReplaceAllString(name, `${1} ${2}`)
	name = strings.NewReplacer("_", " ", "-", " ").Replace(name)
	words := strings.Fields(name)
	for index, word := range words {
		runes := []rune(strings.ToLower(word))
		if len(runes) > 0 {
			runes[0] = unicode.ToUpper(runes[0])
			words[index] = string(runes)
		}
	}
	return strings.Join(words, " ")
}

var mapCamelCaseBoundary = regexp.MustCompile(`([a-z0-9])([A-Z])`)

func numberToInt32(value any) (int32, bool) {
	switch typed := value.(type) {
	case int:
		return int32(typed), true
	case int32:
		return typed, true
	case int64:
		return int32(typed), true
	case float64:
		return int32(typed), true
	case json.Number:
		parsed, err := typed.Int64()
		if err != nil {
			return 0, false
		}
		return int32(parsed), true
	default:
		return 0, false
	}
}

var (
	mapPoolRegistryMu sync.RWMutex
	mapPoolRegistry   = map[string]*MapPoolStore{}
)

// MapPools returns the pool store of a game, opening it on first use. It
// returns nil when the game has not been initialized.
func MapPools(gameId string) *MapPoolStore {
	mapPoolRegistryMu.RLock()
	store := mapPoolRegistry[gameId]
	mapPoolRegistryMu.RUnlock()
	if store != nil {
		return store
	}
	mapPoolRegistryMu.Lock()
	defer mapPoolRegistryMu.Unlock()
	if store = mapPoolRegistry[gameId]; store != nil {
		return store
	}
	store = NewMapPoolStoreForGame(gameId)
	mapPoolRegistry[gameId] = store
	return store
}

// registerMapPools installs the store used by a freshly created game.
func registerMapPools(gameId string, store *MapPoolStore) {
	mapPoolRegistryMu.Lock()
	defer mapPoolRegistryMu.Unlock()
	mapPoolRegistry[gameId] = store
}

// ResetMapPoolRegistryForTests drops every cached store. Tests that point
// userDataFolder at a temporary directory call it between cases.
func ResetMapPoolRegistryForTests() {
	mapPoolRegistryMu.Lock()
	defer mapPoolRegistryMu.Unlock()
	for _, store := range mapPoolRegistry {
		_ = store.Close()
	}
	mapPoolRegistry = map[string]*MapPoolStore{}
}
